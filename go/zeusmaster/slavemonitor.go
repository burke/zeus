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
	slog "github.com/burke/zeus/go/shinylog"
)

type SlaveMonitor struct {
	tree *ProcessTree
	booted chan string
}

func StartSlaveMonitor(tree *ProcessTree, local *net.UnixConn, remote *os.File, quit chan bool) {
	monitor := &SlaveMonitor{tree, make(chan string)}

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
		case node := <- monitor.tree.Dead:
			monitor.slaveDidDie(node)
		}
	}
}

func (mon *SlaveMonitor) cleanupChildren() {
	killSlave(mon.tree.Root)
}

func (mon *SlaveMonitor) slaveDidBoot(slaveName string) {
	bootedSlave := mon.tree.FindSlaveByName(slaveName)
	slog.SlaveBooted(bootedSlave.Name)
	for _, slave := range bootedSlave.Slaves {
		go mon.bootSlave(slave)
	}
}

func (mon *SlaveMonitor) slaveDidDie(slave *SlaveNode) {
	slave.Wipe()
	go mon.bootSlave(slave)
}

func killSlave(slave *SlaveNode) {
	slave.mu.Lock()
	defer slave.mu.Unlock()

	slave.Wipe()

	for _, s := range slave.Slaves {
		go killSlave(s)
	}
}

func (mon *SlaveMonitor) bootSlave(slave *SlaveNode) {
	slave.Parent.WaitUntilBooted()
	msg := CreateSpawnSlaveMessage(slave.Name)
	slave.Parent.mu.Lock()
	slave.Parent.Socket.Write([]byte(msg))
	slave.Parent.mu.Unlock()
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

	go slaveNode.Run(identifier, pid, slaveSocket, mon.booted)
}

