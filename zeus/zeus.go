package zeus

import (
	"syscall"
	"strings"
	"os"
	"time"
	"os/exec"
	"fmt"
	"net"
	"errors"
)

func Run (conf Config) (error) {
	masterSockLocal, masterSockRemote, err := socketpair(syscall.SOCK_DGRAM)
	if err != nil {
		return err
	}

	masterUSockLocal, err := makeUnixSocket(masterSockLocal)
	if err != nil {
		return err
	}

	go runInitialCommand(masterSockRemote, conf.Command)
	go slaveRegistrationHandler(masterUSockLocal)

	time.Sleep(500 * time.Millisecond)

	return nil
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
		fd, err := readFileDescriptorFromUnixSocket(sock)
		if err != nil {
			fmt.Println(err)
		}
		clientFile := fdToFile(fd, "boot-sock")
		clientSocket, err := makeUnixSocket(clientFile)
		if err != nil {
			fmt.Println(err)
		}

		go handleSlaveRegistration(clientSocket)
	}
}

func handleSlaveRegistration(clientSocket *net.UnixConn) {
	// We now expect the client to use this fd they send us to send a Pid&Identifier Message
	msg, _, err := readFromUnixSocket(clientSocket)
	if err != nil {
		fmt.Println(err)
	}
	pid, identifier, err := parsePidMessage(msg)
	fmt.Println(pid, identifier)
	if err != nil {
		fmt.Println(err)
	}

	// Now that we have that, we look up and send the action for that identifier:
	msg = createActionMessage("File.open('omg.log','a'){|f|f.puts 'HAHA BUSINESS TIME'}")
	clientSocket.Write([]byte(msg))

	// It will respond with its status
	msg, _, err = readFromUnixSocket(clientSocket)
	if err != nil {
		fmt.Println(err)
	}
	msg, err = parseActionResponseMessage(msg)
	if err != nil {
		fmt.Println(err)
	}
	if msg == "OK" {
		msg = createSpawnSlaveMessage("default_bundle")
		clientSocket.Write([]byte(msg))
	} else {
		fmt.Println(errors.New(msg))
	}

	// And now we could instruct it to fork:

}

