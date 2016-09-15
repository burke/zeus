package unixsocket

import (
	"errors"
	"fmt"
	"net"
	"syscall"
)

type oobReader struct {
	Conn *net.UnixConn
	oob  []byte
	oobs [][]byte
}

func (o *oobReader) ReadFD() (int, error) {
	if len(o.oobs) > 0 {
		oob := o.oobs[0]
		o.oobs = o.oobs[1:]
		return extractFileDescriptorFromOOB(oob)
	}

	b := make([]byte, 0)
	_, err := o.Read(b)
	if err != nil {
		return 0, err
	}

	if len(o.oobs) > 0 {
		oob := o.oobs[0]
		o.oobs = o.oobs[1:]
		return extractFileDescriptorFromOOB(oob)
	}

	return 0, errors.New("No FD received :(")
}

func (o *oobReader) Read(b []byte) (int, error) {
	n, oobn, _, _, err := o.Conn.ReadMsgUnix(b, o.oob)
	if oobn > 0 {
		newOob := make([]byte, oobn)
		copy(newOob, o.oob[:oobn])
		o.oobs = append(o.oobs, newOob)
	}
	// ReadMsgUnix can return a -1 byte count on error which violate
	// the expectations of io.Reader
	if n == -1 {
		n = 0
	}
	return n, err
}

func extractFileDescriptorFromOOB(oob []byte) (int, error) {
	scms, err := syscall.ParseSocketControlMessage(oob)
	if err != nil {
		return 0, err
	}
	if len(scms) != 1 {
		return 0, errors.New(fmt.Sprintf("expected 1 SocketControlMessage; got scms = %#v", scms))
	}
	scm := scms[0]
	gotFds, err := syscall.ParseUnixRights(&scm)
	if err != nil {
		return 0, err
	}
	if len(gotFds) != 1 {
		return 0, errors.New(fmt.Sprintf("wanted 1 fd; got %#v", gotFds))
	}
	return gotFds[0], nil
}
