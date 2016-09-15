package node

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/burke/zeus/go/processtree/process"
	slog "github.com/burke/zeus/go/shinylog"
)

type State string

const (
	SUnbooted State = "U"
	SBooting  State = "B"
	SReady    State = "R"
	SCrashed  State = "!C"
)

type Node struct {
	bootNode  func() (process.NodeProcess, error)
	watchFile func(string)
	name      string
	changed   chan struct{}
	stop      chan struct{}
	requests  chan process.BootRequest
}

func NewNode(
	name string,
	bootNode func() (process.NodeProcess, error),
	watchFile func(string),
) *Node {
	return &Node{
		bootNode:  bootNode,
		watchFile: watchFile,
		name:      name,
		changed:   make(chan struct{}, 1),
		stop:      make(chan struct{}),
		requests:  make(chan process.BootRequest),
	}
}

func (n *Node) FileChanged() {
	select {
	case n.changed <- struct{}{}:
	default:
	}
}

func (n *Node) Stop() {
	close(n.stop)
	for req := range n.requests {
		req.Error(process.ErrProcessStopping)
	}
}

func (n *Node) Run(state chan<- State) {
	defer func() { close(state) }()
	for {
		n.trace(nil, "entering state SUnbooted")
		state <- SUnbooted
		proc, err := n.bootNode()

		if err == nil {
			n.trace(proc, "initialized from parent %d, entering state SBooting", proc.ParentPid())
			state <- SBooting

			procStop := make(chan struct{})
			go func(files <-chan string) {
				for f := range files {
					n.watchFile(f)
				}
				select {
				case <-procStop:
					// Exiting normally
				default:
					n.trace(proc, "file pipe closed, process likely died")
				}
			}(proc.Files())

			select {
			case <-n.stop:
			case err = <-proc.Errors():
			case <-proc.Ready():
				n.trace(proc, "entering state SReady")
				state <- SReady
				err = n.runProcess(proc)
			}

			close(procStop)
			if stopErr := proc.Stop(); stopErr != nil {
				n.trace(proc, "error stopping process: %v", stopErr)
			}
		}

		if err != nil {
			n.trace(proc, "entering state SCrashed")
			state <- SCrashed
			n.handleCrash(err)
		}

		// TODO: Debounce changes here
		select {
		case <-n.stop:
			n.trace(proc, "stopping node")
			close(n.requests)
			return
		default:
		}
	}
}

func (n *Node) runProcess(proc process.NodeProcess) error {
	for {
		select {
		case err := <-proc.Errors():
			return err
		case <-n.stop:
			return nil
		case <-n.changed:
			return nil
		case req := <-n.requests:
			n.trace(proc, "sending boot request %s", req)
			proc.Boot(req)
		}
	}
}

func (n *Node) handleCrash(err error) {
	for {
		select {
		case <-n.stop:
			return
		case <-n.changed:
			return
		case req := <-n.requests:
			req.Error(err)
		}
	}
}

func (n *Node) BootNode(identifier string) (process.NodeProcess, error) {
	for {
		req := process.NewNodeRequest(identifier)

		// TODO: Address race with closing requests channel
		select {
		case <-n.stop:
			req.Error(process.ErrProcessStopping)
		case n.requests <- req.BootRequest:
		}

		proc, err := req.Await()
		// If the process stopped before it could handle the request,
		// attempt to enqueue it again for the next process to handle.
		if err == process.ErrProcessStopping {
			continue
		}
		if err != nil {
			return nil, err
		}

		if proc.Identifier() != identifier {
			proc.Stop()
			return nil, fmt.Errorf("booted node with identifier %s but expected %s", proc.Identifier, identifier)
		}
		return proc, nil
	}
}

func (n *Node) BootCommand(identifier string, client process.CommandClient) (process.CommandProcess, error) {
	for {
		req := process.NewCommandRequest(identifier, client)

		// TODO: Address race with closing requests channel
		select {
		case <-n.stop:
			req.Error(process.ErrProcessStopping)
		case n.requests <- req.BootRequest:
		}

		proc, err := req.Await()
		// If the process stopped before it could handle the request,
		// attempt to enqueue it again for the next process to handle.
		if err == process.ErrProcessStopping {
			continue
		}
		if err != nil {
			return nil, err
		}

		return proc, nil
	}
}

func (n *Node) trace(proc process.NodeProcess, format string, args ...interface{}) {
	if !slog.TraceEnabled() {
		return
	}

	_, file, line, _ := runtime.Caller(1)

	idx := strings.LastIndex(file, "/go/")
	if idx != -1 {
		file = file[idx+4:]
	} else {
		_, file = filepath.Split(file)
	}

	var prefix string
	if proc != nil {
		prefix = fmt.Sprintf("[%s:%d] %s/(%d) ", file, line, n.name, proc.Pid())
	} else {
		prefix = fmt.Sprintf("[%s:%d] %s/(no PID) ", file, line, n.name)
	}

	slog.Trace(prefix+format, args...)
}
