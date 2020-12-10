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

type Usock struct {
	reader *oobReader
	rbuf   *bufio.Reader

	sync.Mutex
}

func New(conn *net.UnixConn) *Usock {
	u := &Usock{
		reader: &oobReader{
			oob:  make([]byte, 32),
			Conn: conn,
		},
	}
	u.rbuf = bufio.NewReader(u.reader)
	return u
}

func NewFromFile(f *os.File) (*Usock, error) {
	fileConn, err := net.FileConn(f)
	if err != nil {
		return nil, err
	}

	unixConn, ok := fileConn.(*net.UnixConn)
	if !ok {
		return nil, fmt.Errorf("unexpected FileConn type; expected UnixConn, got %T", unixConn)
	}

	return New(unixConn), nil
}

func (u *Usock) Close() {
	u.reader.Conn.Close()
}

func (u *Usock) ReadMessage() (s string, err error) {
	u.Lock()
	defer u.Unlock()

	for {
		s, err = u.rbuf.ReadString(0)
		if err == nil {
			s = strings.TrimRight(s, "\000")
		}
		if err != nil || s != "" {
			return
		}
	}
}

func (u *Usock) WriteMessage(msg string) (int, error) {
	u.Lock()
	defer u.Unlock()

	completeMessage := strings.NewReader(msg + "\000")
	n, err := io.Copy(u.reader.Conn, completeMessage)
	return int(n - 1), err
}

func (u *Usock) ReadFD() (int, error) {
	u.Lock()
	defer u.Unlock()

	return u.reader.ReadFD()
}

func (u *Usock) WriteFD(fd int) error {
	u.Lock()
	defer u.Unlock()

	rights := syscall.UnixRights(fd)

	dummyByte := []byte{0}
	n, oobn, err := u.reader.Conn.WriteMsgUnix(dummyByte, rights, nil)
	if err != nil {
		str := fmt.Sprintf("Usock#WriteFD:WriteMsgUnix: %v / %v\n", err, syscall.EINVAL)
		return errors.New(str)
	}
	if n != 1 || oobn != len(rights) {
		str := fmt.Sprintf("Usock#WriteFD:WriteMsgUnix = %d, %d; want 1, %d\n", n, oobn, len(rights))
		return errors.New(str)
	}
	return nil
}
