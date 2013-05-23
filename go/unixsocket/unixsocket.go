package unixsocket

import (
	"bufio"
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

func Socketpair(typ int) (a, b *os.File, err error) {
	fd, err := syscall.Socketpair(syscall.AF_UNIX, typ, 0)
	if err != nil {
		e := os.NewSyscallError("socketpair", err.(syscall.Errno))
		return nil, nil, e
	}

	a = os.NewFile(uintptr(fd[0]), "socketpair-a")
	b = os.NewFile(uintptr(fd[1]), "socketpair-b")
	return
}

type Usock struct {
	Conn *net.UnixConn
	oobs [][]byte

	sync.Mutex
}

func New(conn *net.UnixConn) *Usock {
	return &Usock{Conn: conn}
}

func NewFromFile(f *os.File) (*Usock, error) {
	fileConn, err := net.FileConn(f)
	if err != nil {
		return nil, err
	}

	unixConn, ok := fileConn.(*net.UnixConn)
	if !ok {
		return nil, errors.New(fmt.Sprintf("unexpected FileConn type; expected UnixConn, got %T", unixConn))
	}

	return New(unixConn), nil
}

func (usock *Usock) Close() {
	usock.Conn.Close()
}

func (u *Usock) ReadFD() (int, error) {
	u.Lock()
	defer u.Unlock()

	if len(u.oobs) > 0 {
		oob := u.oobs[0]
		u.oobs = u.oobs[1:]
		return extractFileDescriptorFromOOB(oob)
	}

	b := make([]byte, 0)
	_, err := u.readLocked(b)
	if err != nil {
		return -1, err
	}

	if len(u.oobs) > 0 {
		oob := u.oobs[0]
		u.oobs = u.oobs[1:]
		return extractFileDescriptorFromOOB(oob)
	}

	return -1, errors.New("No FD received :(")
}

func (usock *Usock) ReadMessage() (s string, err error) {
	r := bufio.NewReader(usock)
	s, err = r.ReadString(0)
	if err == nil {
		s = strings.TrimRight(s, "\000")
	}
	return
}

func (u *Usock) Read(b []byte) (int, error) {
	u.Lock()
	defer u.Unlock()
	return u.readLocked(b)
}

func (usock *Usock) WriteMessage(msg string) (int, error) {
	completeMessage := strings.NewReader(msg + "\000")

	n, err := io.Copy(usock.Conn, completeMessage)
	return int(n - 1), err
}

func (usock *Usock) WriteFD(fd int) error {
	rights := syscall.UnixRights(fd)

	dummyByte := []byte{0}
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
func (u *Usock) readLocked(b []byte) (int, error) {
	oob := make([]byte, 32)
	n, oobn, _, _, err := u.Conn.ReadMsgUnix(b, oob)
	if oobn > 0 {
		newOob := make([]byte, oobn)
		copy(newOob, oob[:oobn])
		u.oobs = append(u.oobs, newOob)
	}
	return n, err
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
