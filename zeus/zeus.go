package zeus

import (
	"syscall"
	"os"
	"encoding/json"
	"os/exec"
	"fmt"
	"errors"
)

type StageBootInfo struct {
	Pid int
	Identifier string
}

func runInitialCommand(sock *os.File) {
	cmd := exec.Command("/Users/burke/.rbenv/shims/ruby", "/Users/burke/go/src/github.com/burke/zeus/a.rb")
	cmd.Env = append(os.Environ(), fmt.Sprintf("ZEUS_MASTER_FD=%d", sock.Fd()))
	cmd.ExtraFiles = []*os.File{sock}

	go cmd.Run()
}

func Run () (error) {
	masterSockLocal, masterSockRemote, err := socketpair(syscall.SOCK_DGRAM)
	if err != nil {
		return err
	}

	masterUSockLocal, err := makeUnixSocket(masterSockLocal)
	if err != nil {
		return err
	}

	runInitialCommand(masterSockRemote)

	// Having just started the process, we expect an IO, which we convert to a UNIX domain socket
	_, fd, err := readFromUnixSocket(masterUSockLocal)
	if err != nil {
		return err
	}
	if fd < 0 {
		return errors.New("expected file descriptor, but got none")
	}
	clientFile := fdToFile(fd, "boot-sock")
	clientSocket, err := makeUnixSocket(clientFile)
	if err != nil {
		return err
	}

	// We now expect the client to use this fd they send us to send a JSON-encoded representation of their PID and identifier...
	msg, _, err := readFromUnixSocket(clientSocket)
	if err != nil {
		return err
	}
	var bootInfo StageBootInfo
	err = json.Unmarshal([]byte(msg), &bootInfo)
	if err != nil {
		return err
	}

	// Now that we have that, we look up and send the action for that identifier:
	clientSocket.Write([]byte("File.open('omg.log','a'){|f|f.puts 'HAHA BUSINESS TIME'}"))

	// It will respond with its status
	msg, _, err = readFromUnixSocket(clientSocket)
	if err != nil {
		return err
	}
	if msg == "OK" {
		clientSocket.Write([]byte("default_bundle"))
	} else {
		return errors.New(msg)
	}

	// And now we could instruct it to fork:

	return nil
}

