// package unixsocket implements type Usock which supports out-of-band transfer
// of open files. The API is {Read,Write}{FD,Message}. Messages are buffered in
// both directions.
//
// When ReadFD is called, it will interpret the first queued OOB datum already
// received, if any. Otherwise it will attempt to receive OOB data from the next
// packet. The packet will be assumed to be a UnixRights message, granting
// access to an open file from another process, and will be decoded as such.
//
// ReadFD will not always block if called in a scenario where there is pending
// data but no OOB data in the first pending packet. This situation is undefined
// (realistically, currently un-thought-about, as zeus has a very regular
// protocol that obviates this concern).
package unixsocket

import (
	"os"
	"syscall"
)

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

var sockName string

func init() {
	sockName = os.Getenv("ZEUSSOCK")
	if sockName == "" {
		sockName = ".zeus.sock"
	}
}

func ZeusSockName() string {
	return sockName
}
