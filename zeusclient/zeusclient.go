package zeusclient

import (
	"fmt"
	"os"
	"syscall"
	"errors"
	"unsafe"
	"net"
	"strings"
	"strconv"

	"github.com/kr/pty"
	usock "github.com/burke/zeus/unixsocket"
)

const (
	sys_TIOCGETA = 0x40487413
	sys_TIOCSETA = 0x80487414
	sys_ISTRIP = 0x20
	sys_INLCR = 0x40
	sys_ICRNL = 0x100
	sys_IGNCR = 0x80
	sys_IXON = 0x200
	sys_IXOFF = 0x400
	sys_ECHO = 0x8
	sys_ICANON = 0x100
	sys_ISIG = 0x80
	termios_NCCS = 20

	zeusSockName = ".zeus.sock"
)

type tcflag_t uint64 // unsigned long
type cc_t byte       // unsigned char
type speed_t uint64  // unsigned long

type struct_termios struct {
	c_iflag tcflag_t       /* input flags */
	c_oflag tcflag_t       /* output flags */
	c_cflag tcflag_t       /* control flags */
	c_lflag tcflag_t       /* local flags */
	c_cc[termios_NCCS] cc_t /* control chars */
	c_ispeed speed_t      /* input speed */
	c_ospeed speed_t      /* output speed */
}


func Run() {
	master, slave, err := pty.Open()
	if err != nil {
		panic(err)
	}
	defer master.Close()
	defer slave.Close()

	if isTerminal(os.Stdout.Fd()) {
		makeTerminalRaw(os.Stdout.Fd())
	}

	mirrorWinsize(os.Stdout, master)

	addr, err := net.ResolveUnixAddr("unixgram", zeusSockName)
	if err != nil {
		panic("Can't resolve server address")
	}

	// TODO: WINCH
	conn, err := net.DialUnix("unix", nil, addr)
	if err != nil {
		panic("Can't connect to Master")
	}

	msg := "Q:console:[]\n"
	conn.Write([]byte(msg))

	usock.SendFdOverUnixSocket(conn, int(slave.Fd()))
	slave.Close()

	msg, _, err = usock.ReadFromUnixSocket(conn)
	if err != nil {
		panic(err)
	}

	parts := strings.Split(msg, "\n")
	pid, err := strconv.Atoi(parts[0])
	if err != nil {
		panic(err)
	}

	println("PID:", pid)
	var exitStatus int = -1
	if len(parts) > 2 {
		exitStatus, err = strconv.Atoi(parts[0])
		if err != nil {
			panic(err)
		}
	}

	for {
		buf := make([]byte,1024)
		n, err := master.Read(buf)
		if err != nil {
			break
		}
		os.Stdout.Write(buf[:n])
	}


	if exitStatus == -1 {
		msg, _, err = usock.ReadFromUnixSocket(conn)
		if err != nil {
			panic(err)
		}
		parts := strings.Split(msg, "\n")
		exitStatus, err = strconv.Atoi(parts[0])
		if err != nil {
			panic(err)
		}
	}

	os.Exit(exitStatus)

}

func isTerminal(fd uintptr) bool {
	var termios struct_termios
	_, _, err := syscall.Syscall6(syscall.SYS_IOCTL, fd, uintptr(sys_TIOCGETA), uintptr(unsafe.Pointer(&termios)), 0, 0, 0)
	return err == 0
}


func mirrorWinsize(from, to *os.File) error {
	var n int
	err := ioctl(from.Fd(), syscall.TIOCGWINSZ, uintptr(unsafe.Pointer(&n)))
	if err != nil {
		return err
	}
	err = ioctl(to.Fd(), syscall.TIOCSWINSZ, uintptr(unsafe.Pointer(&n)))
	if err != nil {
		return err
	}
	return nil
}

func ioctl(fd uintptr, cmd uintptr, ptr uintptr) error {
	_, _, e := syscall.Syscall(
		syscall.SYS_IOCTL,
		fd,
		cmd,
		uintptr(unsafe.Pointer(ptr)),
	)
	if e != 0 {
		return errors.New(fmt.Sprintf("ioctl failed! %s", e))
	}
	return nil
}

func makeTerminalRaw(fd uintptr) error {
	var s struct_termios
	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, fd, uintptr(sys_TIOCGETA), uintptr(unsafe.Pointer(&s)), 0, 0, 0); err != 0 {
		return err
	}

	s.c_iflag &^= sys_ISTRIP | sys_INLCR | sys_ICRNL | sys_IGNCR | sys_IXON | sys_IXOFF
	s.c_lflag &^= sys_ECHO | sys_ICANON | sys_ISIG
	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, fd, uintptr(sys_TIOCSETA), uintptr(unsafe.Pointer(&s)), 0, 0, 0); err != 0 {
		return err
	}

	return nil
}


// https://code.google.com/p/go/source/browse/ssh/terminal/util.go?repo=crypto&r=33d6505b6597ddd49a330ed2f8707bcb2c52318c
/*
func makeTerminalRaw(fd uintptr) error {
	var s syscall.Termios
	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, fd, uintptr(syscall.TCGETS), uintptr(unsafe.Pointer(&s)), 0, 0, 0); err != 0 {
		return err
	}

	s.Iflag &^= syscall.ISTRIP | syscall.INLCR | syscall.ICRNL | syscall.IGNCR | syscall.IXON | syscall.IXOFF
	s.Lflag &^= syscall.ECHO | syscall.ICANON | syscall.ISIG
	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, fd, uintptr(syscall.TCSETS), uintptr(unsafe.Pointer(&s)), 0, 0, 0); err != 0 {
		return err
	}

	return nil
}

*/
