package zeusmaster

import (
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

func StartSlaveMonitor(tree *ProcessTree, local *net.UnixConn, remote *os.File) {
	monitor := &SlaveMonitor{tree: tree}

	monitor.booted = make(chan string)
	monitor.dead = make(chan string)

	go monitor.watchBootedSlaves()
	go monitor.watchDeadSlaves()
	go monitor.watchSlaveRegistrations(local)

	monitor.startInitialProcess(remote)
}

func (mon *SlaveMonitor) watchBootedSlaves() {
	for {
		bootedSlave := mon.tree.FindSlaveByName(<- mon.booted)
		fmt.Println("INFO:", bootedSlave.Name, "is booted.")
		for _, slave := range bootedSlave.Slaves {
			go mon.bootSlave(slave)
		}
	}
}

func (mon *SlaveMonitor) watchDeadSlaves() {

}

func (mon *SlaveMonitor) bootSlave() {

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

func (mon *SlaveMonitor) watchSlaveRegistrations(sock *net.UnixConn) {
	for {
		// Having just started the process, we expect an IO, which we convert to a UNIX domain socket
		fd, err := usock.ReadFileDescriptorFromUnixSocket(sock)
		if err != nil {
			fmt.Println(err)
		}
		clientFile := usock.FdToFile(fd, "boot-sock")
		clientSocket, err := usock.MakeUnixSocket(clientFile)
		if err != nil {
			fmt.Println(err)
		}

		go mon.handleSlaveRegistration(clientSocket)
	}
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
	node.Pid = pid

	if err != nil {
		fmt.Println(err)
	}

	// Now that we have that, we look up and send the action for that identifier:
	msg = CreateActionMessage(node.Action)
	clientSocket.Write([]byte(msg))

	// It will respond with its status
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


