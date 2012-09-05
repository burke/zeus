package zeusmaster

import (
	"net"
	"fmt"
	"syscall"
	"sync"

	usock "github.com/burke/zeus/go/unixsocket"
	slog "github.com/burke/zeus/go/shinylog"
)

type SlaveNode struct {
	ProcessTreeNode
	Socket *net.UnixConn
	Pid int
	Error string
	isBooted bool
	bootWait sync.RWMutex
	Slaves []*SlaveNode
	Commands []*CommandNode
	Features map[string]bool
	ClientCommandPTYFileDescriptor chan int
}

func (node *SlaveNode) Run(identifier string, pid int, slaveSocket *net.UnixConn, booted chan string) {
	// TODO: We actually don't really want to prevent killing this
	// process while it's booting up.
	node.mu.Lock()
	defer node.mu.Unlock()

	node.Pid = pid

	// The slave will execute its action and respond with a status...
	msg, _, err := usock.ReadFromUnixSocket(slaveSocket)
	if err != nil {
		fmt.Println(err)
	}
	msg, err = ParseActionResponseMessage(msg)
	if err != nil {
		fmt.Println(err)
	}
	if msg == "OK" {
		node.Socket = slaveSocket
	} else {
		node.RegisterError(msg)
	}
	node.SignalBooted()
	booted <- identifier

	go node.handleMessages()
}

func (node *SlaveNode) WaitUntilBooted() {
	node.bootWait.RLock()
	node.bootWait.RUnlock()
}

func (node *SlaveNode) SignalBooted() {
	if !node.isBooted {
		node.bootWait.Unlock()
		node.isBooted = true
	}
}

func (node *SlaveNode) SignalUnbooted() {
	if node.isBooted {
		node.bootWait.Lock()
		node.isBooted = false
	}
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

func (node *SlaveNode) Restart(restart chan *SlaveNode) {
	node.mu.Lock()
	defer node.mu.Unlock()

	if processWasKilled := node.tryKillProcess() ; processWasKilled {
		slog.SlaveKilled(node.Name)
		node.SignalUnbooted()
		node.Wipe()
		// if it's not running? I guess... it's starting up already?
		restart <- node
	}

	for _, s := range node.Slaves {
		go s.Restart(restart)
	}
}

// We want to make this the single interface point with the socket.
// we want to republish unneeded messages to channels so other modules
//can pick them up. (notably, clienthandler.)
func (node *SlaveNode) handleMessages() {
	socket := node.Socket
	for {
		if msg, fd, err := usock.ReadFromUnixSocket(socket) ; err != nil {
			node.crashed()
			return
		} else if fd > 0 {
			// File descriptors are sent during client negotiation
			node.ClientCommandPTYFileDescriptor <- fd
		} else {
			// Every other message indicates a feature loaded, and should be sent to filemonitor.
			node.handleFeatureMessage(msg)
		}
	}
}

func (node *SlaveNode) handleFeatureMessage(msg string) {
	if file, err := ParseFeatureMessage(msg) ; err != nil {
		fmt.Println("slavenode.go:handleFeatureMessage:", err)
	} else {
		node.Features[file] = true
		AddFile(file)
	}
}
