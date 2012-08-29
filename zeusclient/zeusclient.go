package zeusclient

import (
	"os"
	"os/signal"
	"net"
	"regexp"
	"strings"
	"strconv"
	"syscall"

	"github.com/kr/pty"
	"github.com/burke/ttyutils"
	slog "github.com/burke/zeus/shinylog"
	usock "github.com/burke/zeus/unixsocket"
)

const (
	zeusSockName = ".zeus.sock"
	sigIntStr = "\x03"
	sigQuitStr = "\x1C"
	sigTstpStr = "\x1A"
)

var signalRegex = regexp.MustCompile(sigIntStr + "|" + sigQuitStr + "|" + sigTstpStr)

func Run(color bool) {
	if !color {
		slog.DisableColor()
		DisableErrorColor()
	}
	master, slave, err := pty.Open()
	if err != nil {
		panic(err)
	}
	defer master.Close()

	if ttyutils.IsTerminal(os.Stdout.Fd()) {
		oldState, err := ttyutils.MakeTerminalRaw(os.Stdout.Fd())
		if err != nil {
			panic(err)
		}
		defer ttyutils.RestoreTerminalState(os.Stdout.Fd(), oldState)
	}

	ttyutils.MirrorWinsize(os.Stdout, master)

	addr, err := net.ResolveUnixAddr("unixgram", zeusSockName)
	if err != nil {
		panic("Can't resolve server address")
	}

	conn, err := net.DialUnix("unix", nil, addr)
	if err != nil {
		ErrorCantConnectToMaster()
	}

	msg := CreateCommandAndArgumentsMessage(os.Args[1], os.Args[2:])
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

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGWINCH, syscall.SIGCONT)
	go func() {
		for sig := range c {
			if sig == syscall.SIGCONT {
				syscall.Kill(pid, syscall.SIGCONT)
			} else if sig == syscall.SIGWINCH {
				ttyutils.MirrorWinsize(os.Stdout, master)
				syscall.Kill(pid, syscall.SIGWINCH)
			}
		}
	}()

	var exitStatus int = -1
	if len(parts) > 2 {
		exitStatus, err = strconv.Atoi(parts[0])
		if err != nil {
			panic(err)
		}
	}

	eof := make(chan bool)
	go func() {
		for {
			buf := make([]byte,1024)
			n, err := master.Read(buf)
			if err != nil {
				eof <- true
				break
			}
			os.Stdout.Write(buf[:n])
		}
	}()

	go func() {
		for {
			buf := make([]byte,1024)
			n, err := os.Stdin.Read(buf)
			if err != nil {
				eof <- true
				break
			}
			// TODO: Since we have a byte array, this could actually just check integer values...
			matches := signalRegex.FindAll(buf, 9999)
			for _, match := range matches {
				if m := string(match); m == sigIntStr {
					syscall.Kill(pid, syscall.SIGINT)
				} else if m == sigQuitStr {
					syscall.Kill(pid, syscall.SIGQUIT)
				} else if m == sigTstpStr {
					syscall.Kill(pid, syscall.SIGTSTP)
					syscall.Kill(os.Getpid(), syscall.SIGTSTP)
				}
			}
			master.Write(buf[:n])
		}
	}()

	<- eof

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

