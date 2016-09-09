package processtree

import (
	"os"
	"strings"
	"sync"
	"time"

	"fmt"
	"runtime"

	"github.com/burke/zeus/go/filemonitor"
	"github.com/burke/zeus/go/processtree/process"
	slog "github.com/burke/zeus/go/shinylog"
	"github.com/burke/zeus/go/unixsocket"
)

const (
	forceKillTimeout = time.Second
)

type SlaveNode struct {
	ProcessTreeNode
	process     process.NodeProcess
	Error       error
	Slaves      []*SlaveNode
	Commands    []*CommandNode
	fileMonitor filemonitor.FileMonitor

	hasSuccessfullyBooted bool

	needsRestart chan bool
	bootRequests chan process.BootRequest

	L        sync.Mutex
	features map[string]bool
	featureL sync.Mutex
	state    string
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
	s.bootRequests = make(chan process.BootRequest)
	s.features = make(map[string]bool)
	s.Name = identifier
	s.Parent = parent
	s.fileMonitor = monitor
	tree.SlavesByName[identifier] = &s
	return &s
}

func (s *SlaveNode) RequestRestart() {
	// Enqueue the restart if there isn't already one in the channel
	select {
	case s.needsRestart <- true:
	default:
	}
}

func (s *SlaveNode) BootNode(identifier string) (process.NodeProcess, error) {
	for {
		req := process.NewNodeRequest(identifier)

		s.bootRequests <- req.BootRequest

		proc, err := req.Await()
		// If the process stopped before it could handle the request,
		// attempt to enqueue it again for the next process to handle.
		if err == process.ErrProcessStopping {
			continue
		}
		if err != nil {
			return nil, err
		}

		if proc.Name() != identifier {
			proc.Stop()
			return nil, fmt.Errorf("booted node with identifier %s but expected %s", proc.Name, identifier)
		}
		return proc, nil
	}
}

func (s *SlaveNode) BootCommand(identifier string) (*unixsocket.Usock, error) {
	for {
		req := process.NewCommandRequest(identifier)

		s.bootRequests <- req.BootRequest

		sock, err := req.Await()
		// If the process stopped before it could handle the request,
		// attempt to enqueue it again for the next process to handle.
		if err == process.ErrProcessStopping {
			continue
		}
		if err != nil {
			return nil, err
		}

		return sock, nil
	}
}

func (s *SlaveNode) Run(changes chan<- bool, command string) {
	nextState := SUnbooted
	for {
		s.L.Lock()
		s.state = nextState
		s.L.Unlock()
		changes <- true
		switch nextState {
		case SUnbooted:
			s.trace("entering state SUnbooted")
			nextState = s.doUnbootedState(command)
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
func (s *SlaveNode) doUnbootedState(command string) string { // -> {SBooting, SCrashed}
	var proc process.NodeProcess
	var err error
	if s.Parent == nil {
		parts := strings.Split(command, " ")
		proc, err = process.StartProcess(parts)
	} else {
		proc, err = s.Parent.BootNode(s.Name)
	}

	s.L.Lock()
	defer s.L.Unlock()
	if err != nil {
		s.Error = err
		return SCrashed
	}

	s.process = proc

	return SBooting
}

// In "SBooting", we have a pid and socket to the process we will use,
// but it has not yet finished initializing (generally, running the code
// specific to this slave). When we receive a message about the success or
// failure of this operation, we transition to either crashed or ready.
func (s *SlaveNode) doBootingState() string { // -> {SCrashed, SReady}
	go func(files <-chan string) {
		for f := range files {
			s.featureL.Lock()
			s.features[f] = true
			s.featureL.Unlock()
			s.fileMonitor.Add(f)
		}
	}(s.process.Files())

	select {
	case err := <-s.process.Errors():
		if stopErr := s.process.Stop(); stopErr != nil {
			s.trace("error stopping process: %v", stopErr)
		}

		s.L.Lock()
		defer s.L.Unlock()
		s.Error = err
		return SCrashed
	case <-s.process.Ready():
		return SReady
	}
}

// In the "SReady" state, we have a functioning process we can spawn
// new processes of of. We respond to requests to boot slaves and
// run commands until we receive a request to restart. This kills
// the process and transitions to SUnbooted.
func (s *SlaveNode) doReadyState() string { // -> SUnbooted
	s.hasSuccessfullyBooted = true

	// If we have a queued restart, service that rather than booting
	// slaves or commands on potentially stale code.
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
		case req := <-s.bootRequests:
			s.trace("sending boot request %s", req)
			s.process.Boot(req)
		}
	}
}

// In the "SCrashed" state, we have an error message from starting
// a process to propogate to the user and all slave nodes. We will
// continue propogating the error until we receive a request to restart.
func (s *SlaveNode) doCrashedState() string { // -> SUnbooted
	// If we have a queued restart, service that rather than booting
	// slaves or commands on potentially stale code.
	select {
	case <-s.needsRestart:
		s.doRestart()
		return SUnbooted
	default:
	}

	// If this is not an error propogated from a parent, output
	// it in the trace logs
	if _, ok := s.Error.(parentError); !ok {
		s.trace("crashed with an error: %s", s.Error)
	}

	for {
		select {
		case <-s.needsRestart:
			s.doRestart()
			return SUnbooted
		case req := <-s.bootRequests:
			req.Error(parentError(s.Error.Error()))
		}
	}
}

func (s *SlaveNode) doRestart() {
	for _, slave := range s.Slaves {
		slave.RequestRestart()
	}
}

func (s *SlaveNode) ForceKill() {
	if s.process != nil {
		s.process.Stop()
	}
}

func (s *SlaveNode) wipe() {
	s.Error = nil
	s.process = nil
}

func (s *SlaveNode) trace(format string, args ...interface{}) {
	if !slog.TraceEnabled() {
		return
	}

	_, file, line, _ := runtime.Caller(1)

	var prefix string
	if s.process != nil {
		prefix = fmt.Sprintf("[%s:%d] %s/(%d)", file, line, s.Name, s.process.Pid)
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

type parentError string

func (e parentError) Error() string {
	return string(e)
}
