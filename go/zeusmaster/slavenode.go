package zeusmaster

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"syscall"

	slog "github.com/burke/zeus/go/shinylog"
	"github.com/burke/zeus/go/unixsocket"
)

type SlaveNode struct {
	ProcessTreeNode
	socket                *unixsocket.Usock
	Pid                   int
	Error                 string
	Slaves                []*SlaveNode
	Commands              []*CommandNode
	Features              map[string]bool
	featureHandlerRunning bool

	needsRestart        chan bool            // size 1
	commandBootRequests chan *CommandRequest // size 256
	slaveBootRequests   chan *SlaveNode          // size 256

	L           sync.Mutex
	state       string
	stateChange *sync.Cond

	event chan bool
}

const (
	sWaiting  = "W"
	sUnbooted = "U"
	sBooting  = "B"
	sReady    = "R"
	sCrashed  = "C"
)

func (tree *ProcessTree) NewSlaveNode(identifier string, parent *SlaveNode) *SlaveNode {
	s := SlaveNode{}
	s.needsRestart = make(chan bool, 1)
	s.commandBootRequests = make(chan *CommandRequest, 256)
	s.slaveBootRequests = make(chan *SlaveNode, 256)
	s.Features = make(map[string]bool)
	s.event = make(chan bool, 1)
	s.Name = identifier
	s.Parent = parent
	var mutex sync.Mutex
	s.stateChange = sync.NewCond(&mutex)
	tree.SlavesByName[identifier] = &s
	return &s
}

func (s *SlaveNode) WaitUntilReadyOrCrashed() {
	s.stateChange.L.Lock()
	for s.state != sReady && s.state != sCrashed && len(s.needsRestart) == 0 {
		s.stateChange.Wait()
	}
	s.stateChange.L.Unlock()
}

// If the slave is executing, or has executed its action, trigger a restart.
// There's no need to trigger a restart if the slave has not yet begun to
// execute its action, and there's no need to queue multiple restarts.
func (s *SlaveNode) RequestRestart() {
	s.L.Lock()
	if s.state == sBooting || s.state == sReady || s.state == sCrashed {
		if len(s.needsRestart) == 0 {
			s.needsRestart <- true
		}
	}
	s.L.Unlock()
	for _, slave := range s.Slaves {
		slave.RequestRestart()
	}
}

func (s *SlaveNode) RequestSlaveBoot(slave *SlaveNode) {
	s.L.Lock()
	s.slaveBootRequests <- slave
	s.L.Unlock()
}

func (s *SlaveNode) RequestCommandBoot(request *CommandRequest) {
	s.L.Lock()
	s.commandBootRequests <- request
	s.L.Unlock()
}

func (s *SlaveNode) SlaveWasInitialized(pid int, usock *unixsocket.Usock) {
	s.L.Lock()
	s.wipe()
	s.Pid = pid
	s.socket = usock
	if s.state == sUnbooted {
		s.event <- true
	} else {
		if pid > 0 {
			syscall.Kill(pid, syscall.SIGKILL)
		}
		slog.ErrorString("Unexpected process for slave `" + s.Name + "` was killed")
	}
	s.L.Unlock()
}

func (s *SlaveNode) Run(monitor *SlaveMonitor) {
	nextState := sWaiting
	for {
		s.L.Lock()
		s.changeState(nextState)
		s.L.Unlock()
		switch nextState {
		case sWaiting:
			nextState = s.doWaitingState()
		case sUnbooted:
			nextState = s.doUnbootedState(monitor)
		case sBooting:
			nextState = s.doBootingState()
		case sCrashed:
			nextState = s.doCrashedOrReadyState()
		case sReady:
			nextState = s.doCrashedOrReadyState()
		default:
			slog.FatalErrorString("Unrecognized state: " + nextState)
		}
	}
}

// These "doXState" functions are called when a SlaveNode enters a state. They are expected
// to continue to execute until 

// The "sWaiting" state represents the state where a Slave is currently
// not running, and neither is its parent. Before we can start booting
// this slave, we must first wait for its parent to finish booting, so
// that we can fork off of it.
func (s *SlaveNode) doWaitingState() string { // -> sUnbooted
	if s.Parent == nil {
		// this is the root state. We get to skip this step. Hooray!
		return sUnbooted
	}
	s.Parent.WaitUntilReadyOrCrashed()
	return sUnbooted
}

// "sUnbooted" represents the state where the parent process is ready, but
// we do not yet have the PID of a process to use for *this* node. In this
// state, we tell the parent process to spawn a process for us, and wait
// to hear back from the SlaveMonitor.
func (s *SlaveNode) doUnbootedState(monitor *SlaveMonitor) string { // -> {sBooting, sCrashed}
	if s.Parent == nil {
		s.L.Lock()
		parts := strings.Split(monitor.tree.ExecCommand, " ")
		cmd := exec.Command(parts[0], parts[1:]...)
		file := monitor.remoteMasterFile
		cmd.Env = append(os.Environ(), fmt.Sprintf("ZEUS_MASTER_FD=%d", file.Fd()))
		cmd.ExtraFiles = []*os.File{file}
		go babysitRootProcess(cmd)
		s.L.Unlock()
	} else {
		s.Parent.RequestSlaveBoot(s)
	}
	<-s.event // sent by StateWasInitialized
	s.L.Lock()
	defer s.L.Unlock()
	if s.Error != "" {
		return sCrashed
	}
	return sBooting
}

// In "sBooting", we have a pid and socket to the process we will use,
// but it has not yet finished initializing (generally, running the code
// specific to this slave. When we receive a message about the success or
// failure of this operation, we transition to either crashed or ready.
func (s *SlaveNode) doBootingState() string { // -> {sCrashed, sReady}
	// The slave will execute its action and respond with a status...
	// Note we don't hold the mutex while waiting for the action to execute.
	msg, err := s.socket.ReadMessage()
	if err != nil {
		slog.Error(err)
	}

	s.L.Lock()
	defer s.L.Unlock()

	msg, err = ParseActionResponseMessage(msg)
	if err != nil {
		slog.ErrorString("[" + s.Name + "] " + err.Error())
	}
	if msg == "OK" {
		return sReady
	}

	if s.Pid > 0 {
		syscall.Kill(s.Pid, syscall.SIGKILL)
	}
	s.wipe()
	s.Error = msg
	return sCrashed
}

// In the "sCrashed" and "sReady" states, we have either a functioning process
// we can spawn new processes off of, or an error message to propagate to
// the user. The high-level operation of these two states is identical:
// First, we work off the queue of command and slave boot requests that have
// built up while this process was booting. Then, we begin a 3-way select
// over those channels with the addition of the "restart" channel, which
// kills the process and transitions us to "sWaiting".
// In this way, we always serve queued fork requests before killing the process.
func (s *SlaveNode) doCrashedOrReadyState() string { // -> sWaiting
	s.L.Lock()
	if s.state == sReady && !s.featureHandlerRunning {
		s.featureHandlerRunning = true
		go s.handleMessages()
	}
	s.L.Unlock()

	s.bootQueuedCommandsAndSlaves()

	for {
		select {
		case slave := <-s.slaveBootRequests:
			s.L.Lock()
			s.bootSlave(slave)
			s.L.Unlock()
		case request := <-s.commandBootRequests:
			s.L.Lock()
			s.bootCommand(request)
			s.L.Unlock()
		case <-s.needsRestart:
			s.L.Lock()
			s.ForceKill()
			s.wipe()
			s.L.Unlock()
			return sWaiting
		}
	}
	return "impossible"
}

// This should only be called while holding a lock on s.L.
func (s *SlaveNode) bootSlave(slave *SlaveNode) {
	if s.Error != "" {
		slave.L.Lock()
		slave.Error = s.Error
		slave.event <- true
		slave.L.Unlock()
		return
	}
	msg := CreateSpawnSlaveMessage(slave.Name)
	_, err := s.socket.WriteMessage(msg)
	if err != nil {
		slog.Error(err)
	}
}

// This should only be called while holding a lock on s.L.
// This unfortunately holds the mutex for a little while, and if the
// command dies super early, the entire slave pretty well deadlocks.
// TODO: review this.
func (s *SlaveNode) bootCommand(request *CommandRequest) {
	identifier := request.Name
	// TODO: If crashed, do something different...
	msg := CreateSpawnCommandMessage(identifier)
	_, err := s.socket.WriteMessage(msg)
	if err != nil {
		slog.Error(err)
		return
	}
	commandFD, err := s.socket.ReadFD()
	if err != nil {
		fmt.Println(s.socket)
		slog.Error(err)
		return
	}
	fileName := strconv.Itoa(rand.Int())
	commandFile := unixsocket.FdToFile(commandFD, fileName)
	request.Retchan <- commandFile
}

func (s *SlaveNode) bootQueuedCommandsAndSlaves() {
	s.L.Lock()
	for i := 0; i < len(s.commandBootRequests); i += 1 {
		request := <-s.commandBootRequests
		s.bootCommand(request)
	}
	for i := 0; i < len(s.slaveBootRequests); i += 1 {
		slave := <-s.slaveBootRequests
		s.bootSlave(slave)
	}
	s.L.Unlock()
}

// This should only be called while holding a lock on s.L.
func (s *SlaveNode) changeState(newState string) {
	StatusUpdate(s.Name, newState)
	// slog.Yellow(newState + " : " + s.Name)
	s.stateChange.L.Lock()
	s.state = newState
	s.stateChange.Broadcast()
	s.stateChange.L.Unlock()
}

func (s *SlaveNode) ForceKill() {
	// note that we don't try to lock the mutex.
	if s.Pid > 0 {
		syscall.Kill(s.Pid, syscall.SIGKILL)
	}
}

func (s *SlaveNode) wipe() {
	s.Pid = 0
	s.socket = nil
	s.Error = ""
}

func babysitRootProcess(cmd *exec.Cmd) {
	// We want to let this process run "forever", but it will eventually
	// die... either on program termination or when its dependencies change
	// and we kill it. when it's requested to restart, err is "signal 9",
	// and we do nothing.
	output, err := cmd.CombinedOutput()
	if err == nil {
		ErrorConfigCommandCrashed(string(output))
	}
	msg := err.Error()
	if len(msg) > 11 && err.Error()[:11] != "exit status" {
		ErrorConfigCommandCouldntStart(err.Error())
	}
}

// We want to make this the single interface point with the socket.
// we want to republish unneeded messages to channels so other modules
//can pick them up. (notably, clienthandler.)
func (s *SlaveNode) handleMessages() {
	for {
		if msg, err := s.socket.ReadMessage(); err != nil {
			s.L.Lock()
			s.featureHandlerRunning = false
			s.L.Unlock()
			return
		} else {
			msg = strings.TrimRight(msg, "\000")
			s.handleFeatureMessage(msg)
		}
	}
}

func (s *SlaveNode) handleFeatureMessage(msg string) {
	if file, err := ParseFeatureMessage(msg); err != nil {
		slog.Error(err)
	} else {
		s.Features[file] = true
		AddFile(file)
	}
}
