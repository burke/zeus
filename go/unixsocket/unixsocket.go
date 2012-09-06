package unixsocket

import (
	"syscall"
	"strings"
	"os"
	"errors"
	"net"
	"fmt"
)

// http://code.google.com/p/rsc/source/browse/fuse/mount_linux.go
// https://github.com/hanwen/go-fuse/blob/master/fuse/mount.go
// http://code.google.com/p/go/source/browse/src/pkg/syscall/syscall_bsd.go?spec=svn982df2b2cb4b6001e8b60f9e8a000751e9a42198&name=982df2b2cb4b&r=982df2b2cb4b6001e8b60f9e8a000751e9a42198

type Usock struct {
	Conn *net.UnixConn
}

func NewUsock(conn *net.UnixConn) *Usock {
	return &Usock{conn}
}

func FdToFile(fd int, name string) (*os.File) {
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

func (usock *Usock) WriteFD(fd int) error {
	rights := syscall.UnixRights(fd)

	dummyByte := []byte("x")
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

func (usock *Usock) ReadFD() (fd int, err error) {
	_, fd, err = usock.ReadMessage()
	if err != nil && fd < 0 {
		err = errors.New("Invalid File Descriptor")
	}
	return
}

func (usock *Usock) WriteMessage(msg string) (int, error) {
	return usock.Conn.Write([]byte(msg + "\000"))
}

func (usock *Usock) Close() {
	usock.Conn.Close()
}

// func ReadFromUnixSocket(sock *net.UnixConn) (msg string, fd int, err error) {
func (usock *Usock) ReadMessage() (msg string, fd int, err error) {
	buf := make([]byte, 1024) // if FD: 1 byte   ; else: varies
	oob := make([]byte, 32)   // if FD: 24 bytes ; else: 0

	n, oobn, _, _, err := usock.Conn.ReadMsgUnix(buf, oob)
	if err != nil {
		return "", -1, err
	}
	if oobn == 0 {
		msg := strings.TrimRight(string(buf[:n]), "\000")
		return msg, -1, nil
	} else {
		if fd, err := extractFileDescriptorFromOOB(oob[:oobn]) ; err != nil {
			return "", -1, err
		} else {
			return "", fd, nil
		}
	}
	return "", -1, nil // just to satisfy the stupid compiler...
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
