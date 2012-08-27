package zeusmaster

import (
	"syscall"
	"strings"
	"os"
	"os/exec"
	"fmt"
	"net"
	"errors"

	usock "github.com/burke/zeus/unixsocket"
)

type SlaveMonitor struct {
	tree *ProcessTree
	booted chan string
	dead   chan string
}

func StartSlaveMonitor(tree *ProcessTree, local *net.UnixConn, remote *os.File, quit chan bool) {
	monitor := &SlaveMonitor{tree, make(chan string), make(chan string)}

	// We just want this unix socket to be a channel so we can select on it...
	registeringFds := make(chan int, 3)
	go func() {
		for {
			fd, err := usock.ReadFileDescriptorFromUnixSocket(local)
			if err != nil {
				fmt.Println(err)
			}
			registeringFds <- fd
		}
	}()

	go monitor.startInitialProcess(remote)

	for {
		select {
		case <- quit:
			quit <- true
			monitor.cleanupChildren()
			return
		case fd := <- registeringFds:
			monitor.slaveDidBeginRegistration(fd)
		case name := <- monitor.booted:
			monitor.slaveDidBoot(name)
		case name := <- monitor.dead:
			monitor.slaveDidDie(name)
		}
	}
}

func (mon *SlaveMonitor) cleanupChildren() {
	println("Cleaning up slaves")
	killSlave(mon.tree.Root)
}

func (mon *SlaveMonitor) slaveDidBoot(slaveName string) {
	bootedSlave := mon.tree.FindSlaveByName(slaveName)
	fmt.Println("INFO:", bootedSlave.Name, "is booted.")
	for _, slave := range bootedSlave.Slaves {
		go mon.bootSlave(slave)
	}
}

func (mon *SlaveMonitor) slaveDidDie(slaveName string) {
	deadSlave := mon.tree.FindSlaveByName(slaveName)
	go killSlave(deadSlave)
}

func killSlave(slave *SlaveNode) {
	slave.mu.Lock()
	defer slave.mu.Unlock()

	pid := slave.Pid
	fmt.Println("INFO:", slave.Name, "is being killed.")
	println("Cleaning up slaves", pid)
	syscall.Kill(pid, 9) // error implies already dead -- no worries.
	slave.Pid = 0
	slave.Socket = nil

	for _, s := range slave.Slaves {
		go killSlave(s)
	}
}

func (mon *SlaveMonitor) bootSlave(slave *SlaveNode) {
	if slave.Parent.Pid < 1 {
		panic("Can't boot a slave with an unbooted parent")
	}
	msg := CreateSpawnSlaveMessage(slave.Name)
	slave.Parent.Socket.Write([]byte(msg))
}

func (mon *SlaveMonitor) startInitialProcess(sock *os.File) {
	command := mon.tree.ExecCommand
	parts := strings.SplitN(command, " ", 2)
	executable := parts[0]
	args := parts[1]
	cmd := exec.Command(executable, args)
	cmd.Env = append(os.Environ(), fmt.Sprintf("ZEUS_MASTER_FD=%d", sock.Fd()))
	cmd.ExtraFiles = []*os.File{sock}

	cmd.Run()
}

func (mon *SlaveMonitor) slaveDidBeginRegistration(fd int) {
	// Having just started the process, we expect an IO, which we convert to a UNIX domain socket
	clientFile := usock.FdToFile(fd, "boot-sock")
	clientSocket, err := usock.MakeUnixSocket(clientFile)
	if err != nil {
		fmt.Println(err)
	}

	go mon.handleSlaveRegistration(clientSocket)
}

func (mon *SlaveMonitor) handleSlaveRegistration(clientSocket *net.UnixConn) {
	// We now expect the client to use this fd they send us to send a Pid&Identifier Message
	msg, _, err := usock.ReadFromUnixSocket(clientSocket)
	if err != nil {
		fmt.Println(err)
	}
	pid, identifier, err := ParsePidMessage(msg)

	node := mon.tree.FindSlaveByName(identifier)
	if node == nil {
		panic("Unknown identifier")
	}

	// TODO: We actually don't really want to prevent killing this
	// process while it's booting up.
	node.mu.Lock()
	defer node.mu.Unlock()

	node.Pid = pid

	if err != nil {
		fmt.Println(err)
	}

	// The slave will execute its action and respond with a status...
	msg, _, err = usock.ReadFromUnixSocket(clientSocket)
	if err != nil {
		fmt.Println(err)
	}
	msg, err = ParseActionResponseMessage(msg)
	if err != nil {
		fmt.Println(err)
	}
	if msg == "OK" {
		node.Socket = clientSocket
		mon.booted <- identifier
	} else {
		// TODO: handle failed boots.
		fmt.Println(errors.New(msg))
	}

}
