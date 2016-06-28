package processtree

import (
	"bufio"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"fmt"
	"runtime"

	"github.com/burke/zeus/go/filemonitor"
	"github.com/burke/zeus/go/messages"
	slog "github.com/burke/zeus/go/shinylog"
	"github.com/burke/zeus/go/unixsocket"
)

type SlaveNode struct {
	ProcessTreeNode
	socket      *unixsocket.Usock
	pid         int
	Error       string
	Slaves      []*SlaveNode
	Commands    []*CommandNode
	fileMonitor filemonitor.FileMonitor

	hasSuccessfullyBooted bool

	needsRestart        chan bool
	slaveBootRequests   chan *SlaveNode
	commandBootRequests chan *CommandRequest

	L        sync.Mutex
	features map[string]bool
	featureL sync.Mutex
	state    string

	event chan bool
}

type CommandReply struct {
	State string
	File  *os.File
}

type CommandRequest struct {
	Name    string
	Retchan chan *CommandReply
}

const (
	SUnbooted = "U"
	SBooting  = "B"
	SReady    = "R"
	SCrashed  = "C"
)

func (tree *ProcessTree) NewSlaveNode(identifier string, parent *SlaveNode, monitor filemonitor.FileMonitor) *SlaveNode {
	s := SlaveNode{}
	s.needsRestart = make(chan bool, 1)
	s.slaveBootRequests = make(chan *SlaveNode, 256)
	s.commandBootRequests = make(chan *CommandRequest, 256)
	s.features = make(map[string]bool)
	s.event = make(chan bool)
	s.Name = identifier
	s.Parent = parent
	s.fileMonitor = monitor
	tree.SlavesByName[identifier] = &s
	return &s
}

func (s *SlaveNode) RequestRestart() {
	s.L.Lock()
	defer s.L.Unlock()

	// If this slave is currently waiting on process to boot,
	// unhang it. If it isn't, the error will get wiped anyway.
	s.Error = "Received restart request while booting"
	s.ReportBootEvent()

	// Enqueue the restart if there isn't already one in the channel
	select {
	case s.needsRestart <- true:
	default:
	}
}

func (s *SlaveNode) RequestSlaveBoot(slave *SlaveNode) {
	s.slaveBootRequests <- slave
}

func (s *SlaveNode) RequestCommandBoot(request *CommandRequest) {
	s.commandBootRequests <- request
}

func (s *SlaveNode) ReportBootEvent() bool {
	select {
	case s.event <- true:
		return true
	default:
		return false
	}
}

func (s *SlaveNode) SlaveWasInitialized(pid int, usock *unixsocket.Usock, featurePipeFd int) {
	file := os.NewFile(uintptr(featurePipeFd), "featurepipe")

	s.L.Lock()
	if !s.ReportBootEvent() {
		if pid > 0 {
			syscall.Kill(pid, syscall.SIGKILL)
		}
		slog.ErrorString(fmt.Sprintf("Unexpected process %d for slave %q was killed", pid, s.Name))
	} else {
		s.wipe()
		s.pid = pid
		s.socket = usock
		go s.handleMessages(file)
		s.trace("initialized slave %s with pid %d", s.Name, pid)
	}
	s.L.Unlock()
}

func (s *SlaveNode) Run(monitor *SlaveMonitor) {
	nextState := SUnbooted
	for {
		s.L.Lock()
		s.state = nextState
		s.L.Unlock()
		monitor.tree.StateChanged <- true
		switch nextState {
		case SUnbooted:
			nextState = s.doUnbootedState(monitor)
		case SBooting:
			nextState = s.doBootingState()
		case SReady:
			nextState = s.doReadyState()
		case SCrashed:
			nextState = s.doCrashedState()
		default:
			slog.FatalErrorString("Unrecognized state: " + nextState)
		}
	}
}

func (s *SlaveNode) State() string {
	s.L.Lock()
	defer s.L.Unlock()

	return s.state
}

func (s *SlaveNode) HasFeature(file string) bool {
	s.featureL.Lock()
	defer s.featureL.Unlock()
	return s.features[file]
}

// These "doXState" functions are called when a SlaveNode enters a state. They are expected
// to continue to execute until

// "SUnbooted" represents the state where we do not yet have the PID
// of a process to use for *this* node. In this state, we wait for the
// parent process to spawn a process for us and hear back from the
// SlaveMonitor.
func (s *SlaveNode) doUnbootedState(monitor *SlaveMonitor) string { // -> {SBooting, SCrashed}
	s.trace("in unbooted state")
	if s.Parent == nil {
		s.L.Lock()
		parts := strings.Split(monitor.tree.ExecCommand, " ")
		cmd := exec.Command(parts[0], parts[1:]...)
		file := monitor.remoteMasterFile
		cmd.Env = append(os.Environ(), fmt.Sprintf("ZEUS_MASTER_FD=%d", file.Fd()))
		cmd.ExtraFiles = []*os.File{file}
		go s.babysitRootProcess(cmd)
		s.L.Unlock()
	} else {
		s.Parent.RequestSlaveBoot(s)
	}

	<-s.event // sent by SlaveWasInitialized

	s.L.Lock()
	defer s.L.Unlock()
	if s.Error != "" {
		return SCrashed
	}
	return SBooting
}

// In "SBooting", we have a pid and socket to the process we will use,
// but it has not yet finished initializing (generally, running the code
// specific to this slave). When we receive a message about the success or
// failure of this operation, we transition to either crashed or ready.
func (s *SlaveNode) doBootingState() string { // -> {SCrashed, SReady}
	// The slave will execute its action and respond with a status...
	// Note we don't hold the mutex while waiting for the action to execute.
	msg, err := s.socket.ReadMessage()
	if err != nil {
		slog.Error(err)
	}
	s.trace("in booting state")
	s.L.Lock()
	defer s.L.Unlock()

	msg, err = messages.ParseActionResponseMessage(msg)
	if err != nil {
		slog.ErrorString("[" + s.Name + "] " + err.Error())
	}
	if msg == "OK" {
		return SReady
	}

	// Clean up:
	if s.pid > 0 {
		syscall.Kill(s.pid, syscall.SIGKILL)
	}
	s.wipe()
	s.Error = msg
	return SCrashed
}

// In the "SReady" state, we have a functioning process we can spawn
// new processes of of. We respond to requests to boot slaves and
// run commands until we receive a request to restart. This kills
// the process and transitions to SUnbooted.
func (s *SlaveNode) doReadyState() string { // -> SUnbooted
	s.hasSuccessfullyBooted = true
	s.trace("entered state SReady")

	for {
		select {
		case <-s.needsRestart:
			s.L.Lock()
			s.ForceKill()
			s.wipe()
			s.L.Unlock()

			for _, slave := range s.Slaves {
				slave.RequestRestart()
			}
			return SUnbooted
		case slave := <-s.slaveBootRequests:
			s.bootSlave(slave)
		case request := <-s.commandBootRequests:
			s.bootCommand(request)
		}
	}
}

// In the "SCrashed" state, we have an error message from starting
// a process to propogate to the user and all slave nodes. We will
// continue propogating the error until we receive a request to restart.
func (s *SlaveNode) doCrashedState() string { // -> SUnbooted
	for {
		select {
		case <-s.needsRestart:
			s.L.Lock()
			s.ForceKill()
			s.wipe()
			s.L.Unlock()
			return SUnbooted
		case slave := <-s.slaveBootRequests:
			slave.L.Lock()
			slave.Error = s.Error
			slave.ReportBootEvent()
			slave.L.Unlock()
		case request := <-s.commandBootRequests:
			s.L.Lock()
			s.trace("reporting crash to command %v", request)
			request.Retchan <- &CommandReply{SCrashed, nil}
			s.L.Unlock()
		}
	}
}

func (s *SlaveNode) bootSlave(slave *SlaveNode) {
	s.L.Lock()
	defer s.L.Unlock()

	s.trace("now sending slave boot request for %s", slave.Name)

	msg := messages.CreateSpawnSlaveMessage(slave.Name)
	_, err := s.socket.WriteMessage(msg)
	if err != nil {
		slog.Error(err)
	}
}

// This unfortunately holds the mutex for a little while, and if the
// command dies super early, the entire slave pretty well deadlocks.
// TODO: review this.
func (s *SlaveNode) bootCommand(request *CommandRequest) {
	s.L.Lock()
	defer s.L.Unlock()

	s.trace("now sending command boot request %v", request)

	identifier := request.Name
	msg := messages.CreateSpawnCommandMessage(identifier)
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
	commandFile := os.NewFile(uintptr(commandFD), fileName)
	request.Retchan <- &CommandReply{s.state, commandFile}
}

func (s *SlaveNode) ForceKill() {
	// note that we don't try to lock the mutex.
	if s.pid > 0 {
		syscall.Kill(s.pid, syscall.SIGKILL)
	}
}

func (s *SlaveNode) wipe() {
	s.pid = 0
	s.socket = nil
	s.Error = ""
}

func (s *SlaveNode) babysitRootProcess(cmd *exec.Cmd) {
	// We want to let this process run "forever", but it will eventually
	// die... either on program termination or when its dependencies change
	// and we kill it. when it's requested to restart, err is "signal 9",
	// and we do nothing.
	s.trace("running the root command now")
	output, err := cmd.CombinedOutput()
	if err == nil {
		// TODO
		s.trace("root process exited; output was: %s", output)
		println(string(output))
		/* ErrorConfigCommandCrashed(string(output)) */
	}
	msg := err.Error()
	if s.hasSuccessfullyBooted == false {
		// TODO
		s.trace("root process exited with an error before it could boot: %s; output was: %s", msg, output)
		println(msg)
		/* ErrorConfigCommandCouldntStart(msg, string(output)) */
	} else if msg == "signal 9" {
		s.trace("root process exited because we killed it & it will be restarted: %s; output was: %s", msg, output)
	} else {
		s.L.Lock()
		defer s.L.Unlock()

		s.trace("root process exited with error. Sending it to crashed state. Message was: %s; output: %s", msg, output)
		s.Error = fmt.Sprintf("Zeus root process (%s) died with message %s:\n%s", s.Name, msg, output)
		if !s.ReportBootEvent() {
			s.trace("Unexpected state for root process to be in at this time: %s", s.state)
		}
	}
}

// We want to make this the single interface point with the socket.
// we want to republish unneeded messages to channels so other modules
//can pick them up. (notably, clienthandler.)
func (s *SlaveNode) handleMessages(featurePipe *os.File) {
	reader := bufio.NewReader(featurePipe)
	for {
		if msg, err := reader.ReadString('\n'); err != nil {
			return
		} else {
			msg = strings.TrimRight(msg, "\n")
			s.featureL.Lock()
			s.features[msg] = true
			s.featureL.Unlock()
			s.fileMonitor.Add(msg)
		}
	}
}

func (s *SlaveNode) trace(format string, args ...interface{}) {
	if !slog.TraceEnabled() {
		return
	}

	_, file, line, _ := runtime.Caller(1)

	var prefix string
	if s.pid != 0 {
		prefix = fmt.Sprintf("[%s:%d] %s/(%d)", file, line, s.Name, s.pid)
	} else {
		prefix = fmt.Sprintf("[%s:%d] %s/(no PID)", file, line, s.Name)
	}
	new_args := make([]interface{}, len(args)+1)
	new_args[0] = prefix
	for i, v := range args {
		new_args[i+1] = v
	}
	slog.Trace("%s "+format, new_args...)
}
