package unixsocket

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"syscall"
)

// http://code.google.com/p/rsc/source/browse/fuse/mount_linux.go
// https://github.com/hanwen/go-fuse/blob/master/fuse/mount.go
// http://code.google.com/p/go/source/browse/src/pkg/syscall/syscall_bsd.go?spec=svn982df2b2cb4b6001e8b60f9e8a000751e9a42198&name=982df2b2cb4b&r=982df2b2cb4b6001e8b60f9e8a000751e9a42198
var sockName string = "NOTSET"

type Usock struct {
	Conn           *net.UnixConn
	mu             sync.Mutex
	readFDs        []int
	readMessages   []string
	partialMessage string
}

func ZeusSockName() string {
  if sockName != "NOTSET" {
    return sockName
  }

  sockName = os.Getenv("ZEUSSOCK")
  if sockName == "" {
    sockName = ".zeus.sock"
  }
  return sockName
}

func NewUsock(conn *net.UnixConn) *Usock {
	return &Usock{Conn: conn}
}

func FdToFile(fd int, name string) *os.File {
	return os.NewFile(uintptr(fd), name)
}

func Socketpair(typ int) (a, b *os.File, err error) {
	fd, err := syscall.Socketpair(syscall.AF_UNIX, typ, 0)
	if err != nil {
		e := os.NewSyscallError("socketpair", err.(syscall.Errno))
		return nil, nil, e
	}

	a = FdToFile(fd[0], "socketpair-a")
	b = FdToFile(fd[1], "socketpair-b")
	return
}

func NewUsockFromFile(f *os.File) (*Usock, error) {
	fileConn, err := net.FileConn(f)
	if err != nil {
		return nil, err
	}

	unixConn, ok := fileConn.(*net.UnixConn)
	if !ok {
		return nil, errors.New(fmt.Sprintf("unexpected FileConn type; expected UnixConn, got %T", unixConn))
	}

	return NewUsock(unixConn), nil
}

func MakeUnixSocket(f *os.File) (*net.UnixConn, error) {
	fileConn, err := net.FileConn(f)
	if err != nil {
		return nil, err
	}

	unixConn, ok := fileConn.(*net.UnixConn)
	if !ok {
		return nil, errors.New(fmt.Sprintf("unexpected FileConn type; expected UnixConn, got %T", unixConn))
	}

	return unixConn, nil
}

func (usock *Usock) Close() {
	usock.Conn.Close()
}

func (usock *Usock) WriteMessage(msg string) (int, error) {
	completeMessage := strings.NewReader(msg + "\000")

	n, err := io.Copy(usock.Conn, completeMessage)
	return int(n - 1), err
}

func (usock *Usock) WriteFD(fd int) error {
	rights := syscall.UnixRights(fd)

	dummyByte := []byte("\000")
	n, oobn, err := usock.Conn.WriteMsgUnix(dummyByte, rights, nil)
	if err != nil {
		str := fmt.Sprintf("Usock#WriteFD:WriteMsgUnix: %v %v\n", err, syscall.EINVAL)
		return errors.New(str)
	}
	if n != 1 || oobn != len(rights) {
		str := fmt.Sprintf("Usock#WriteFD:WriteMsgUnix = %d, %d; want 1, %d\n", n, oobn, len(rights))
		return errors.New(str)
	}
	return nil
}

func (usock *Usock) ReadFD() (int, error) {
	usock.mu.Lock()
	defer usock.mu.Unlock()

	if fd := usock.returnFDIfRead(); fd >= 0 {
		return fd, nil
	}

	if err := usock.readFromSocket(); err != nil {
		return -1, err
	}

	if fd := usock.returnFDIfRead(); fd >= 0 {
		return fd, nil
	}

	return -1, errors.New("Expected File Descriptor from socket; none received.")
}

func (usock *Usock) ReadMessage() (string, error) {
	usock.mu.Lock()
	defer usock.mu.Unlock()

	if msg := usock.returnMessageIfRead(); msg != "" {
		return msg, nil
	}

	if err := usock.readFromSocket(); err != nil {
		return "", err
	}

	if msg := usock.returnMessageIfRead(); msg != "" {
		return msg, nil
	}

	return "", errors.New("Expected message from socket; none received.")
}

func (usock *Usock) ReadMessageOrFD() (string, int, error) {
	usock.mu.Lock()
	defer usock.mu.Unlock()

	if fd := usock.returnFDIfRead(); fd >= 0 {
		return "", fd, nil
	}
	if msg := usock.returnMessageIfRead(); msg != "" {
		return msg, -1, nil
	}

	if err := usock.readFromSocket(); err != nil {
		return "", -1, err
	}

	if fd := usock.returnFDIfRead(); fd >= 0 {
		return "", fd, nil
	}
	if msg := usock.returnMessageIfRead(); msg != "" {
		return msg, -1, nil
	}

	return "", -1, errors.New("Expected message or FD from socket; none received.")
}

func (usock *Usock) returnFDIfRead() int {
	if len(usock.readFDs) > 0 {
		fd := usock.readFDs[0]
		usock.readFDs = usock.readFDs[1:]
		return fd
	}
	return -1
}

func (usock *Usock) returnMessageIfRead() string {
	if len(usock.readMessages) > 0 {
		msg := usock.readMessages[0]
		usock.readMessages = usock.readMessages[1:]
		return msg
	}
	return ""
}

func (usock *Usock) readFromSocket() (err error) {
	buf := make([]byte, 1024)
	oob := make([]byte, 32)

	n, oobn, _, _, err := usock.Conn.ReadMsgUnix(buf, oob)
	if err != nil {
		return err
	}
	if oobn > 0 { // we got a file descriptor.
		if fd, err := extractFileDescriptorFromOOB(oob[:oobn]); err != nil {
			return err
		} else {
			usock.readFDs = append(usock.readFDs, fd)
		}
	}
	if n > 0 {
		// This relies on the fact that a message should be null terminated.
		// `messages` for a single full message ("a message\000") then will be ["a message", ""]
		messages := strings.Split(string(buf[:n]), "\000")
		for index, message := range messages {
			if message == "" {
				continue
			}
			if usock.partialMessage != "" {
				message = usock.partialMessage + message
				usock.partialMessage = ""
			}
			if index == len(messages)-1 {
				usock.partialMessage = message
			} else {
				usock.readMessages = append(usock.readMessages, message)
			}
		}
		// if we only got a partial message, and there's nothing currently buffered to return, read again.
		if len(usock.readMessages) == 0 && usock.partialMessage != "" {
			return usock.readFromSocket()
		}
	}
	return nil
}

func extractFileDescriptorFromOOB(oob []byte) (int, error) {
	scms, err := syscall.ParseSocketControlMessage(oob)
	if err != nil {
		return -1, err
	}
	if len(scms) != 1 {
		return -1, errors.New(fmt.Sprintf("expected 1 SocketControlMessage; got scms = %#v", scms))
	}
	scm := scms[0]
	gotFds, err := syscall.ParseUnixRights(&scm)
	if err != nil {
		return -1, err
	}
	if len(gotFds) != 1 {
		return -1, errors.New(fmt.Sprintf("wanted 1 fd; got %#v", gotFds))
	}
	return gotFds[0], nil
}
