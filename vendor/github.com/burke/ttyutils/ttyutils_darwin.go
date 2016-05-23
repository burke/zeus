package ttyutils

import (
	"errors"
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

const (
	ISTRIP       = 0x20
	INLCR        = 0x40
	ICRNL        = 0x100
	IEXTEN       = 0x400
	ECHOE        = 0x02
	ECHOK        = 0x04
	OPOST        = 0x01
	IMAXBEL      = 0x2000
	IGNBRK       = 0x01
	BRKINT       = 0x02
	IGNCR        = 0x80
	IXON         = 0x200
	IXOFF        = 0x400
	ICANON       = 0x100
	ISIG         = 0x80
	termios_NCCS = 20
)

type tcflag_t uint64 // unsigned long
type cc_t byte       // unsigned char
type speed_t uint64  // unsigned long

type Termios struct {
	Iflag  tcflag_t           /* input flags */
	Oflag  tcflag_t           /* output flags */
	Cflag  tcflag_t           /* control flags */
	Lflag  tcflag_t           /* local flags */
	Cc     [termios_NCCS]cc_t /* control chars */
	Ispeed speed_t            /* input speed */
	Ospeed speed_t            /* output speed */
}

type Ttysize struct {
	Lines   uint16
	Columns uint16
}

func IsTerminal(fd uintptr) bool {
	var termios Termios
	_, _, err := syscall.Syscall6(syscall.SYS_IOCTL, fd, uintptr(syscall.TIOCGETA), uintptr(unsafe.Pointer(&termios)), 0, 0, 0)
	return err == 0
}

func Winsize(of *os.File) (Ttysize, error) {
	var ts Ttysize
	err := ioctl(of.Fd(), syscall.TIOCGWINSZ, uintptr(unsafe.Pointer(&ts)))
	return ts, err
}

func MirrorWinsize(from, to *os.File) error {
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

func NoEcho(fd uintptr) (*Termios, error) {
	var s Termios
	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, fd, uintptr(syscall.TIOCGETA), uintptr(unsafe.Pointer(&s)), 0, 0, 0); err != 0 {
		return nil, err
	}

	oldState := s
	s.Lflag &^= syscall.ECHO | ECHOE | ECHOK
	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, fd, uintptr(syscall.TIOCSETA), uintptr(unsafe.Pointer(&s)), 0, 0, 0); err != 0 {
		return nil, err
	}

	return &oldState, nil
}

func MakeTerminalRaw(fd uintptr) (*Termios, error) {
	var s Termios
	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, fd, uintptr(syscall.TIOCGETA), uintptr(unsafe.Pointer(&s)), 0, 0, 0); err != 0 {
		return nil, err
	}

	oldState := s
	s.Iflag &^= ISTRIP | INLCR | ICRNL | IGNCR | IXON | IXOFF | IMAXBEL | BRKINT
	s.Iflag |= IGNBRK
	s.Lflag &^= syscall.ECHO | ICANON | ISIG | IEXTEN | ECHOE | ECHOK
	s.Oflag &^= OPOST
	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, fd, uintptr(syscall.TIOCSETA), uintptr(unsafe.Pointer(&s)), 0, 0, 0); err != 0 {
		return nil, err
	}

	return &oldState, nil
}

func RestoreTerminalState(fd uintptr, termios *Termios) error {
	_, _, err := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), uintptr(syscall.TIOCSETA), uintptr(unsafe.Pointer(termios)), 0, 0, 0)
	return err
}
