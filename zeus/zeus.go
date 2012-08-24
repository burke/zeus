package zeus

import (
	"syscall"
	"os"
	"encoding/json"
	"os/exec"
	"fmt"
)

type StageBootInfo struct {
	Pid int
	Identifier string
}

func readMessageFromRuby(fd int) (string, error) {
	newFile := fdToFile(fd, "ruby-sock")
	newSocket, _ := makeUnixSocket(newFile)
	msg, _, e := readFromUnixSocket(newSocket)
	if e != nil {
		panic(e)
	}

	var bootInfo StageBootInfo
	err := json.Unmarshal([]byte(msg), &bootInfo)
	if err != nil {
		return "", err
	}

	newSocket.Write([]byte("File.open('omg.log','a'){|f|f.puts 'HAHA BUSINESS'}"))

	msg2, _, _ := readFromUnixSocket(newSocket)
	if msg2 == "OK" {
		newSocket.Write([]byte("default_bundle"))
	}

	return "ya", nil
}

func Run () (string) {
	lf, rf, err := socketpair(syscall.SOCK_DGRAM)
	if err != nil {
		panic(err)
	}

	local, err := makeUnixSocket(lf)
	if err != nil {
		panic(err)
	}

	cmd := exec.Command("/Users/burke/.rbenv/shims/ruby", "/Users/burke/go/src/github.com/burke/zeus/a.rb")
	cmd.Env = append(os.Environ(), fmt.Sprintf("ZEUS_MASTER_FD=%d", rf.Fd()))
	cmd.ExtraFiles = []*os.File{rf}

	go cmd.Run()

	msg, fd, err := readFromUnixSocket(local)
	if err != nil {
		panic(err)
	}
	if fd >= 0 {
		str, _ := readMessageFromRuby(fd)
		return str
	}

	return msg
}

