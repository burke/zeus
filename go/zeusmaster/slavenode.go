package zeusmaster

import (
	"net"
	"fmt"
	"sync"

	usock "github.com/burke/zeus/go/unixsocket"
	slog "github.com/burke/zeus/go/shinylog"
)

type SlaveNode struct {
	ProcessTreeNode
	Socket *net.UnixConn
	Pid int
	Error string
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
	node.bootWait.Unlock()
}

func (node *SlaveNode) SignalUnbooted() {
	node.bootWait.Lock()
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

func (node *SlaveNode) crashed() {
	slog.SlaveDied(node.Name)
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
			node.pushFdMessage(fd)
		} else {
			node.pushMessage(msg)
		}
	}
}

func (node *SlaveNode) pushFdMessage(fd int) {
	node.ClientCommandPTYFileDescriptor <- fd
}

func (node *SlaveNode) pushMessage(message string) {
	println("GOT MESSAGE", message)
}
