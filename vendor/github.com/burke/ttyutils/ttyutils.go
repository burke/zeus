package ttyutils

import (
	"errors"
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

type Termios syscall.Termios

type Ttysize struct {
	Lines   uint16
	Columns uint16
}

func IsTerminal(fd uintptr) bool {
	var termios syscall.Termios
	_, _, err := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), ioctlReadTermios, uintptr(unsafe.Pointer(&termios)), 0, 0, 0)
	return err == 0
}

func Winsize(of *os.File) (Ttysize, error) {
	var dimensions [4]uint16

	err := ioctl(of.Fd(), uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&dimensions)))
	if err != nil {
		return Ttysize{}, err
	}

	return Ttysize{dimensions[0], dimensions[1]}, nil
}

func MirrorWinsize(from, to *os.File) error {
	var dimensions [4]uint16

	err := ioctl(from.Fd(), syscall.TIOCGWINSZ, uintptr(unsafe.Pointer(&dimensions)))
	if err != nil {
		return err
	}
	err = ioctl(to.Fd(), syscall.TIOCSWINSZ, uintptr(unsafe.Pointer(&dimensions)))
	if err != nil {
		return err
	}
	return nil
}

func NoEcho(fd uintptr) (*Termios, error) {
	var s Termios
	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, fd, ioctlReadTermios, uintptr(unsafe.Pointer(&s)), 0, 0, 0); err != 0 {
		return nil, err
	}

	oldState := s
	s.Lflag &^= syscall.ECHO | syscall.ECHOE | syscall.ECHOK
	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, fd, ioctlWriteTermios, uintptr(unsafe.Pointer(&s)), 0, 0, 0); err != 0 {
		return nil, err
	}

	return &oldState, nil
}

func MakeTerminalRaw(fd uintptr) (*Termios, error) {
	var s Termios
	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, fd, ioctlReadTermios, uintptr(unsafe.Pointer(&s)), 0, 0, 0); err != 0 {
		return nil, err
	}

	oldState := s
	s.Iflag &^= syscall.ISTRIP | syscall.INLCR | syscall.ICRNL | syscall.IGNCR | syscall.IXON | syscall.IXOFF | syscall.IMAXBEL | syscall.BRKINT
	s.Iflag |= syscall.IGNBRK
	s.Lflag &^= syscall.ECHO | syscall.ICANON | syscall.ISIG | syscall.IEXTEN | syscall.ECHOE | syscall.ECHOK
	s.Oflag &^= syscall.OPOST
	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, fd, ioctlWriteTermios, uintptr(unsafe.Pointer(&s)), 0, 0, 0); err != 0 {
		return nil, err
	}

	return &oldState, nil
}

func RestoreTerminalState(fd uintptr, termios *Termios) error {
	_, _, err := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(fd), ioctlWriteTermios, uintptr(unsafe.Pointer(termios)), 0, 0, 0)
	return err
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
