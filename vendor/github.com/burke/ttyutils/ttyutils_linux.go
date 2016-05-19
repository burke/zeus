package ttyutils

import (
	"errors"
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

type Termios syscall.Termios

func IsTerminal(fd uintptr) bool {
	var termios syscall.Termios
	_, _, err := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), uintptr(syscall.TCGETS), uintptr(unsafe.Pointer(&termios)), 0, 0, 0)
	return err == 0
}

type Ttysize struct {
	Lines   uint16
	Columns uint16
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
	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, fd, uintptr(syscall.TCGETS), uintptr(unsafe.Pointer(&s)), 0, 0, 0); err != 0 {
		return nil, err
	}

	oldState := s
	s.Lflag &^= syscall.ECHO | syscall.ECHOE | syscall.ECHOK
	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, fd, uintptr(syscall.TCSETS), uintptr(unsafe.Pointer(&s)), 0, 0, 0); err != 0 {
		return nil, err
	}

	return &oldState, nil
}

func MakeTerminalRaw(fd uintptr) (*Termios, error) {
	var s Termios
	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, fd, uintptr(syscall.TCGETS), uintptr(unsafe.Pointer(&s)), 0, 0, 0); err != 0 {
		return nil, err
	}

	oldState := s
	s.Iflag &^= syscall.ISTRIP | syscall.INLCR | syscall.ICRNL | syscall.IGNCR | syscall.IXON | syscall.IXOFF | syscall.IMAXBEL | syscall.BRKINT
	s.Iflag |= syscall.IGNBRK
	s.Lflag &^= syscall.ECHO | syscall.ICANON | syscall.ISIG | syscall.IEXTEN | syscall.ECHOE | syscall.ECHOK
	s.Oflag &^= syscall.OPOST
	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, fd, uintptr(syscall.TCSETS), uintptr(unsafe.Pointer(&s)), 0, 0, 0); err != 0 {
		return nil, err
	}

	return &oldState, nil
}

func RestoreTerminalState(fd uintptr, termios *Termios) error {
	_, _, err := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), uintptr(syscall.TCSETS), uintptr(unsafe.Pointer(termios)), 0, 0, 0)
	return err
}
