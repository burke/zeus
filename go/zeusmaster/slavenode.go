package zeusmaster

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"syscall"

	slog "github.com/burke/zeus/go/shinylog"
	"github.com/burke/zeus/go/unixsocket"
)

type SlaveNode struct {
	ProcessTreeNode
	Socket                         *net.UnixConn
	Pid                            int
	Error                          string
	isBooted                       bool
	bootWait                       *sync.Cond
	restartWait                    *sync.Cond
	Slaves                         []*SlaveNode
	Commands                       []*CommandNode
	Features                       map[string]bool
	ClientCommandPTYFileDescriptor chan int
}

func (tree *ProcessTree) NewSlaveNode(name string, parent *SlaveNode) *SlaveNode {
	x := &SlaveNode{}
	x.Parent = parent
	x.isBooted = false
	x.Name = name
	var mutex sync.Mutex
	x.bootWait = sync.NewCond(&mutex)
	x.restartWait = sync.NewCond(&mutex)
	x.Features = make(map[string]bool)
	x.ClientCommandPTYFileDescriptor = make(chan int)
	tree.SlavesByName[name] = x
	return x
}

func (node *SlaveNode) Run(identifier string, pid int, slaveUsock *unixsocket.Usock) {
	// TODO: We actually don't really want to prevent killing this
	// process while it's booting up.
	node.mu.Lock()
	defer node.mu.Unlock()

	node.Pid = pid

	// The slave will execute its action and respond with a status...
	msg, _, err := slaveUsock.ReadMessage()
	if err != nil {
		fmt.Println(err)
	}
	msg, err = ParseActionResponseMessage(msg)
	if err != nil {
		fmt.Println(err)
	}
	if msg == "OK" {
		node.Socket = slaveUsock.Conn
	} else {
		node.RegisterError(msg)
	}
	node.SignalBooted()
	slog.SlaveBooted(node.Name)

	go node.handleMessages()
}

func (node *SlaveNode) WaitUntilBooted() {
	node.bootWait.L.Lock()
	for !node.isBooted {
		node.bootWait.Wait()
	}
	node.bootWait.L.Unlock()
}

func (node *SlaveNode) WaitUntilUnbooted() {
	node.bootWait.L.Lock()
	for node.isBooted {
		node.bootWait.Wait()
	}
	node.bootWait.L.Unlock()
}

func (node *SlaveNode) SignalBooted() {
	node.bootWait.L.Lock()
	if !node.isBooted {
		node.isBooted = true
		node.bootWait.Broadcast()
	}
	node.bootWait.L.Unlock()
}

func (node *SlaveNode) SignalUnbooted() {
	node.bootWait.L.Lock()
	if node.isBooted {
		node.isBooted = false
		node.bootWait.Broadcast()
	}
	node.bootWait.L.Unlock()
}

func (node *SlaveNode) RequestRestart() {
	node.restartWait.Broadcast()
}

func (node *SlaveNode) WaitUntilRestartRequested() {
	node.restartWait.L.Lock()
	node.restartWait.Wait()
	node.restartWait.L.Unlock()
}

func (node *SlaveNode) RegisterError(msg string) {
	node.Error = msg
	for _, slave := range node.Slaves {
		slave.RegisterError(msg)
	}
}

func (node *SlaveNode) Wipe() {
	node.Pid = 0
	node.Socket = nil
	node.Error = ""
}

// true if process was killed; false otherwise
func (node *SlaveNode) tryKillProcess() bool {
	if node.Pid > 0 {
		err := syscall.Kill(node.Pid, 9)
		return err == nil
	}
	return false
}

func (node *SlaveNode) crashed() {
	node.tryKillProcess()
	// whether or not it was actually dead, our socket was.
	// so just report it as dead already.
	slog.SlaveDied(node.Name)
}

// unceremoniously kill the process. We just need to tidy
// up before program exit.
func (node *SlaveNode) Shutdown() {
	// don't lock the mutex. Just shut down. this is run in a signal handler.
	node.tryKillProcess()
}

func (node *SlaveNode) Kill() {
	node.mu.Lock()
	defer node.mu.Unlock()

	if processWasKilled := node.tryKillProcess(); processWasKilled {
		slog.SlaveKilled(node.Name)
		node.Wipe()
		// TODO: See if this works if not done via goroutine
		go node.SignalUnbooted()
	}
}

// We want to make this the single interface point with the socket.
// we want to republish unneeded messages to channels so other modules
//can pick them up. (notably, clienthandler.)
func (node *SlaveNode) handleMessages() {
	socket := node.Socket
	usock := unixsocket.NewUsock(socket)
	for {
		if msg, fd, err := usock.ReadMessage(); err != nil {
			node.crashed()
			return
		} else if fd > 0 {
			// File descriptors are sent during client negotiation
			node.ClientCommandPTYFileDescriptor <- fd
		} else {
			// Every other message indicates a feature loaded, and should be sent to filemonitor.
			msg = strings.TrimRight(msg, "\000")
			node.handleFeatureMessage(msg)
		}
	}
}

func (node *SlaveNode) handleFeatureMessage(msg string) {
	if file, err := ParseFeatureMessage(msg); err != nil {
		fmt.Println("slavenode.go:handleFeatureMessage:", err)
	} else {
		node.Features[file] = true
		AddFile(file)
	}
}
