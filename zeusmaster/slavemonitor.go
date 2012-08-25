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

func StartSlaveMonitor(tree *ProcessTree) {
	masterSockLocal, masterSockRemote, err := usock.Socketpair(syscall.SOCK_DGRAM)
	if err != nil {
		panic(err)
	}

	masterUSockLocal, err := usock.MakeUnixSocket(masterSockLocal)
	if err != nil {
		panic(err)
	}

	go runInitialCommand(masterSockRemote, tree.ExecCommand)
	go slaveRegistrationHandler(masterUSockLocal)
}

func runInitialCommand(sock *os.File, command string) {
	parts := strings.SplitN(command, " ", 2)
	executable := parts[0]
	args := parts[1]
	cmd := exec.Command(executable, args)
	cmd.Env = append(os.Environ(), fmt.Sprintf("ZEUS_MASTER_FD=%d", sock.Fd()))
	cmd.ExtraFiles = []*os.File{sock}

	cmd.Run()
}

func slaveRegistrationHandler(sock *net.UnixConn) {
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

		go handleSlaveRegistration(clientSocket)
	}
}

func handleSlaveRegistration(clientSocket *net.UnixConn) {
	// We now expect the client to use this fd they send us to send a Pid&Identifier Message
	msg, _, err := usock.ReadFromUnixSocket(clientSocket)
	if err != nil {
		fmt.Println(err)
	}
	pid, identifier, err := ParsePidMessage(msg)
	fmt.Println(pid, identifier)
	if err != nil {
		fmt.Println(err)
	}

	// Now that we have that, we look up and send the action for that identifier:
	msg = CreateActionMessage("File.open('omg.log','a'){|f|f.puts 'HAHA BUSINESS TIME'}")
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
		msg = CreateSpawnSlaveMessage("default_bundle")
		clientSocket.Write([]byte(msg))
	} else {
		fmt.Println(errors.New(msg))
	}

	// And now we could instruct it to fork:

}


