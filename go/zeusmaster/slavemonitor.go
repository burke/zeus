package zeusmaster

import (
	"strconv"
	"math/rand"
	"strings"
	"os"
	"os/exec"
	"fmt"
	"net"

	usock "github.com/burke/zeus/go/unixsocket"
)

type SlaveMonitor struct {
	tree *ProcessTree
}

func StartSlaveMonitor(tree *ProcessTree, local *net.UnixConn, remote *os.File, quit chan bool) {
	monitor := &SlaveMonitor{tree}

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

	for _, slave := range monitor.tree.SlavesByName {
		if slave.Parent == nil {
			go monitor.startInitialProcess(remote)
		} else {
			go monitor.bootSlave(slave)
		}
	}

	for {
		select {
		case <- quit:
			quit <- true
			monitor.cleanupChildren()
			return
		case fd := <- registeringFds:
			monitor.slaveDidBeginRegistration(fd)
		}
	}
}

func (mon *SlaveMonitor) cleanupChildren() {
	for _, slave := range mon.tree.SlavesByName {
		slave.Shutdown()
	}
}

func (mon *SlaveMonitor) bootSlave(slave *SlaveNode) {
	for {
		slave.Parent.WaitUntilBooted()

		msg := CreateSpawnSlaveMessage(slave.Name)
		slave.Parent.Socket.Write([]byte(msg))

		restartNow := make(chan bool)
		go func() {
			slave.Parent.WaitUntilUnbooted()
			restartNow <- true
		}()
		go func() {
			slave.WaitUntilRestartRequested()
			restartNow <- true
		}()

		<- restartNow
		slave.Kill()
	}
}

func (mon *SlaveMonitor) startInitialProcess(sock *os.File) {
	command := mon.tree.ExecCommand
	parts := strings.Split(command, " ")
	executable := parts[0]
	args := parts[1:]
	cmd := exec.Command(executable, args...)
	cmd.Env = append(os.Environ(), fmt.Sprintf("ZEUS_MASTER_FD=%d", sock.Fd()))
	cmd.ExtraFiles = []*os.File{sock}

	// We want to let this process run "forever", but it will eventually
	// die... either on program termination or when its dependencies change
	// and we kill it.
	output, err := cmd.CombinedOutput()
	if err != nil && string(err.Error()[:11]) != "exit status" {
		ErrorConfigCommandCouldntStart(err.Error())
	} else {
		ErrorConfigCommandCrashed(string(output))
	}
}

func (mon *SlaveMonitor) slaveDidBeginRegistration(fd int) {
	// Having just started the process, we expect an IO, which we convert to a UNIX domain socket
	fileName := strconv.Itoa(rand.Int())
	slaveFile := usock.FdToFile(fd, fileName)
	slaveSocket, err := usock.MakeUnixSocket(slaveFile)
	if err != nil {
		fmt.Println(err)
	}
	if err = slaveSocket.SetReadBuffer(262144) ; err != nil {
		fmt.Println(err)
	}
	if err = slaveSocket.SetWriteBuffer(262144) ; err != nil {
		fmt.Println(err)
	}

	// We now expect the slave to use this fd they send us to send a Pid&Identifier Message
	msg, _, err := usock.ReadFromUnixSocket(slaveSocket)
	if err != nil {
		fmt.Println(err)
	}
	pid, identifier, err := ParsePidMessage(msg)

	slaveNode := mon.tree.FindSlaveByName(identifier)
	if slaveNode == nil {
		panic("Unknown identifier")
	}

	go slaveNode.Run(identifier, pid, slaveSocket)
}

