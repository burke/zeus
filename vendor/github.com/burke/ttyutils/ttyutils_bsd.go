// +build darwin dragonfly freebsd netbsd openbsd

package ttyutils

import "syscall"

const ioctlReadTermios = syscall.TIOCGETA
const ioctlWriteTermios = syscall.TIOCSETA
