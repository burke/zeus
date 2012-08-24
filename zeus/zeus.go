package zeus

// http://code.google.com/p/rsc/source/browse/fuse/mount_linux.go
// https://github.com/hanwen/go-fuse/blob/master/fuse/mount.go
// http://code.google.com/p/go/source/browse/src/pkg/syscall/syscall_bsd.go?spec=svn982df2b2cb4b6001e8b60f9e8a000751e9a42198&name=982df2b2cb4b&r=982df2b2cb4b6001e8b60f9e8a000751e9a42198

import (
	"syscall"
	"os"
	"errors"
	"net"
	"os/exec"
	"fmt"
)

func socketpair(typ int) (a, b *os.File, err error) {
	fd, err := syscall.Socketpair(syscall.AF_UNIX, typ, 0)
	if err != nil {
		e := os.NewSyscallError("socketpair", err.(syscall.Errno))
		return nil, nil, e
	}

	a = os.NewFile(uintptr(fd[0]), "socketpair-a")
	b = os.NewFile(uintptr(fd[1]), "socketpair-b")
	return
}

func makeUnixSocket(f *os.File) (*net.UnixConn, error) {
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

func readFromUnixSocket(sock *net.UnixConn) (msg string, fd int, err error) {
	buf := make([]byte, 1024) // if FD: 1 byte   ; else: varies
	oob := make([]byte, 32)   // if FD: 24 bytes ; else: 0

	_, oobn, _, _, err := sock.ReadMsgUnix(buf, oob)
	if oobn == 0 {
		return string(buf), -1, nil
	} else {
		// It's a file descriptor
		scms, err := syscall.ParseSocketControlMessage(oob[:oobn])
		if err != nil {
			return "", -1, err
		}
		if len(scms) != 1 {
			return "", -1, errors.New(fmt.Sprintf("expected 1 SocketControlMessage; got scms = %#v", scms))
		}
		scm := scms[0]
		gotFds, err := syscall.ParseUnixRights(&scm)
		if err != nil {
			return "", -1, err
		}
		if len(gotFds) != 1 {
			return "", -1, errors.New(fmt.Sprintf("wanted 1 fd; got %#v", gotFds))
		}

		return "", gotFds[0], nil
	}
	return "", -1, nil // just to satisfy the stupid compiler...
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

	cmd.Run()

	msg, fd, err := readFromUnixSocket(local)
	if err != nil {
		panic(err)
	}
	if fd >= 0 {
		fl := os.NewFile(uintptr(fd), "a-file")
		nw, _ := makeUnixSocket(fl)
		m, _, e := readFromUnixSocket(nw)
		if e != nil {
			panic(e)
		}
		return m
	}

	return msg
}



