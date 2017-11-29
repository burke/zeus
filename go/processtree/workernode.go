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
	"time"

	"fmt"
	"runtime"

	"github.com/burke/zeus/go/filemonitor"
	"github.com/burke/zeus/go/messages"
	slog "github.com/burke/zeus/go/shinylog"
	"github.com/burke/zeus/go/unixsocket"
)

const (
	forceKillTimeout = time.Second
)

type WorkerNode struct {
	ProcessTreeNode
	socket      *unixsocket.Usock
	pid         int
	Error       string
	Workers     []*WorkerNode
	Commands    []*CommandNode
	fileMonitor filemonitor.FileMonitor

	hasSuccessfullyBooted bool

	needsRestart        chan bool
	commandBootRequests chan *CommandRequest
	workerBootRequests  chan *WorkerNode

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

var humanreadableStates = map[string]string{
	SUnbooted: "unbooted",
	SBooting:  "booted",
	SReady:    "ready",
	SCrashed:  "crashed",
}

func (tree *ProcessTree) NewWorkerNode(identifier string, parent *WorkerNode, monitor filemonitor.FileMonitor) *WorkerNode {
	s := WorkerNode{}
	s.needsRestart = make(chan bool, 1)
	s.workerBootRequests = make(chan *WorkerNode, 256)
	s.commandBootRequests = make(chan *CommandRequest, 256)
	s.features = make(map[string]bool)
	s.event = make(chan bool)
	s.Name = identifier
	s.Parent = parent
	s.fileMonitor = monitor
	tree.WorkersByName[identifier] = &s
	return &s
}

func (s *WorkerNode) RequestRestart() {
	s.L.Lock()
	defer s.L.Unlock()

	// If this worker is currently waiting on a process to boot,
	// unhang it and force it to transition to the crashed state
	// where it will wait for restart messages.
	if s.ReportBootEvent() {
		s.Error = "Received restart request while booting"
	}

	// Enqueue the restart if there isn't already one in the channel
	select {
	case s.needsRestart <- true:
	default:
	}
}

func (s *WorkerNode) RequestWorkerBoot(worker *WorkerNode) {
	s.workerBootRequests <- worker
}

func (s *WorkerNode) RequestCommandBoot(request *CommandRequest) {
	s.commandBootRequests <- request
}

func (s *WorkerNode) ReportBootEvent() bool {
	select {
	case s.event <- true:
		return true
	default:
		return false
	}
}

func (s *WorkerNode) WorkerWasInitialized(pid, parentPid int, usock *unixsocket.Usock, featurePipeFd int) {
	file := os.NewFile(uintptr(featurePipeFd), "featurepipe")

	s.L.Lock()
	if !s.ReportBootEvent() {
		s.forceKillPid(pid)
		s.trace("Unexpected process %d with parent %d for worker %q was killed", pid, parentPid, s.Name)
	} else {
		s.wipe()
		s.pid = pid
		s.socket = usock
		go s.handleMessages(file)
		s.trace("initialized worker %s with pid %d from parent %d", s.Name, pid, parentPid)
	}
	s.L.Unlock()
}

func (s *WorkerNode) Run(monitor *WorkerMonitor) {
	nextState := SUnbooted
	for {
		s.L.Lock()
		s.state = nextState
		s.L.Unlock()
		monitor.tree.StateChanged <- true
		switch nextState {
		case SUnbooted:
			s.trace("entering state SUnbooted")
			nextState = s.doUnbootedState(monitor)
		case SBooting:
			s.trace("entering state SBooting")
			nextState = s.doBootingState()
		case SReady:
			s.trace("entering state SReady")
			nextState = s.doReadyState()
		case SCrashed:
			s.trace("entering state SCrashed")
			nextState = s.doCrashedState()
		default:
			slog.FatalErrorString("Unrecognized state: " + nextState)
		}
	}
}

func (s *WorkerNode) State() string {
	s.L.Lock()
	defer s.L.Unlock()

	return s.state
}

func (s *WorkerNode) HumanReadableState() string {
	return humanreadableStates[s.state]
}

func (s *WorkerNode) HasFeature(file string) bool {
	s.featureL.Lock()
	defer s.featureL.Unlock()
	return s.features[file]
}

// These "doXState" functions are called when a WorkerNode enters a state. They are expected
// to continue to execute until

// "SUnbooted" represents the state where we do not yet have the PID
// of a process to use for *this* node. In this state, we wait for the
// parent process to spawn a process for us and hear back from the
// WorkerMonitor.
func (s *WorkerNode) doUnbootedState(monitor *WorkerMonitor) string { // -> {SBooting, SCrashed}
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
		s.Parent.RequestWorkerBoot(s)
	}

	<-s.event // sent by WorkerWasInitialized

	s.L.Lock()
	defer s.L.Unlock()
	if s.Error != "" {
		return SCrashed
	}
	return SBooting
}

// In "SBooting", we have a pid and socket to the process we will use,
// but it has not yet finished initializing (generally, running the code
// specific to this worker). When we receive a message about the success or
// failure of this operation, we transition to either crashed or ready.
func (s *WorkerNode) doBootingState() string { // -> {SCrashed, SReady}
	// The worker will execute its action and respond with a status...
	// Note we don't hold the mutex while waiting for the action to execute.
	msg, err := s.socket.ReadMessage()
	if err != nil {
		s.L.Lock()
		defer s.L.Unlock()
		s.Error = err.Error()
		slog.ErrorString("[" + s.Name + "] " + err.Error())

		return SCrashed
	}

	s.trace("received action message")
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
// new processes of of. We respond to requests to boot workers and
// run commands until we receive a request to restart. This kills
// the process and transitions to SUnbooted.
func (s *WorkerNode) doReadyState() string { // -> SUnbooted
	s.hasSuccessfullyBooted = true

	// If we have a queued restart, service that rather than booting
	// workers or commands on potentially stale code.
	select {
	case <-s.needsRestart:
		s.doRestart()
		return SUnbooted
	default:
	}

	for {
		select {
		case <-s.needsRestart:
			s.doRestart()
			return SUnbooted
		case worker := <-s.workerBootRequests:
			s.bootWorker(worker)
		case request := <-s.commandBootRequests:
			s.bootCommand(request)
		}
	}
}

// In the "SCrashed" state, we have an error message from starting
// a process to propogate to the user and all worker nodes. We will
// continue propogating the error until we receive a request to restart.
func (s *WorkerNode) doCrashedState() string { // -> SUnbooted
	// If we have a queued restart, service that rather than booting
	// workers or commands on potentially stale code.
	select {
	case <-s.needsRestart:
		s.doRestart()
		return SUnbooted
	default:
	}

	for {
		select {
		case <-s.needsRestart:
			s.doRestart()
			return SUnbooted
		case worker := <-s.workerBootRequests:
			worker.L.Lock()
			worker.Error = s.Error
			worker.ReportBootEvent()
			worker.L.Unlock()
		case request := <-s.commandBootRequests:
			s.L.Lock()
			s.trace("reporting crash to command %v", request)
			request.Retchan <- &CommandReply{SCrashed, nil}
			s.L.Unlock()
		}
	}
}

func (s *WorkerNode) doRestart() {
	s.L.Lock()
	s.ForceKill()
	s.wipe()
	s.L.Unlock()

	// Drain and ignore any enqueued worker boot requests since
	// we're going to make them all restart again anyway.
	drained := false
	for !drained {
		select {
		case <-s.workerBootRequests:
		default:
			drained = true
		}
	}

	for _, worker := range s.Workers {
		worker.RequestRestart()
	}
}

func (s *WorkerNode) bootWorker(worker *WorkerNode) {
	s.L.Lock()
	defer s.L.Unlock()

	s.trace("now sending worker boot request for %s", worker.Name)

	msg := messages.CreateSpawnWorkerMessage(worker.Name)
	_, err := s.socket.WriteMessage(msg)
	if err != nil {
		slog.Error(err)
	}
}

// This unfortunately holds the mutex for a little while, and if the
// command dies super early, the entire worker pretty well deadlocks.
// TODO: review this.
func (s *WorkerNode) bootCommand(request *CommandRequest) {
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

func (s *WorkerNode) ForceKill() {
	// note that we don't try to lock the mutex.
	s.forceKillPid(s.pid)
}

func (s *WorkerNode) wipe() {
	s.pid = 0
	s.socket = nil
	s.Error = ""
}

func (s *WorkerNode) babysitRootProcess(cmd *exec.Cmd) {
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
func (s *WorkerNode) handleMessages(featurePipe *os.File) {
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

func (s *WorkerNode) forceKillPid(pid int) error {
	if pid <= 0 {
		return nil
	}

	if err := syscall.Kill(pid, syscall.SIGTERM); err != nil {
		err = fmt.Errorf("Error killing pid %q: %v", pid, err)
		s.trace(err.Error())
		return err
	}

	exited := make(chan error)
	go func() {
		for {
			if err := syscall.Kill(pid, syscall.Signal(0)); err != nil {
				exited <- nil
				return
			}

			// Since the process is not our direct child, we can't use wait
			// and are forced to poll for completion. We know this won't loop
			// forever because the timeout below will SIGKILL the process
			// which guarantees that it'll go away and we'll get an ESRCH.
			time.Sleep(time.Millisecond)
		}
	}()

	select {
	case err := <-exited:
		if err != nil && err != syscall.ESRCH {
			err = fmt.Errorf("Error sending signal to pid %q: %v", pid, err)
			s.trace(err.Error())
			return err
		}
		return nil
	case <-time.After(forceKillTimeout):
		syscall.Kill(pid, syscall.SIGKILL)
		return nil
	}
}

func (s *WorkerNode) trace(format string, args ...interface{}) {
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
