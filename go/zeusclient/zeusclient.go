package zeusclient

import (
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/burke/pty"
	"github.com/burke/ttyutils"
	slog "github.com/burke/zeus/go/shinylog"
	"github.com/burke/zeus/go/unixsocket"
)

const (
	zeusSockName = ".zeus.sock"
	sigInt       = 3 // todo: this doesn't seem unicode-friendly...
	sigQuit      = 28
	sigTstp      = 26
)

func Run(color bool) {
	os.Exit(doRun(color))
}

func doRun(color bool) int {
	if !color {
		slog.DisableColor()
		DisableErrorColor()
	}

	if os.Getenv("RAILS_ENV") != "" {
		println("Warning: Specifying a Rails environment via RAILS_ENV has no effect for commands run with zeus.")
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
	usock := unixsocket.NewUsock(conn)

	msg := CreateCommandAndArgumentsMessage(os.Args[1], os.Args[2:])
	usock.WriteMessage(msg)
	usock.WriteFD(int(slave.Fd()))
	slave.Close()

	msg, err = usock.ReadMessage()
	if err != nil {
		panic(err)
	}

	parts := strings.Split(msg, "\000")
	commandPid, err := strconv.Atoi(parts[0])
	defer func() {
		if commandPid > 0 {
			// Just in case.
			syscall.Kill(commandPid, 9)
		}
	}()

	if err != nil {
		panic(err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGWINCH, syscall.SIGCONT)
	go func() {
		for sig := range c {
			if sig == syscall.SIGCONT {
				syscall.Kill(commandPid, syscall.SIGCONT)
			} else if sig == syscall.SIGWINCH {
				ttyutils.MirrorWinsize(os.Stdout, master)
				syscall.Kill(commandPid, syscall.SIGWINCH)
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
			buf := make([]byte, 1024)
			n, err := master.Read(buf)
			if err != nil {
				eof <- true
				break
			}
			os.Stdout.Write(buf[:n])
		}
	}()

	go func() {
		buf := make([]byte, 8192)
		for {
			n, err := os.Stdin.Read(buf)
			if err != nil {
				eof <- true
				break
			}
			for i := 0; i < n; i++ {
				switch buf[i] {
				case sigInt:
					syscall.Kill(commandPid, syscall.SIGINT)
				case sigQuit:
					syscall.Kill(commandPid, syscall.SIGQUIT)
				case sigTstp:
					syscall.Kill(commandPid, syscall.SIGTSTP)
					syscall.Kill(os.Getpid(), syscall.SIGTSTP)
				}
			}
			master.Write(buf[:n])
		}
	}()

	<-eof

	if exitStatus == -1 {
		msg, err = usock.ReadMessage()
		if err != nil {
			panic(err)
		}
		parts := strings.Split(msg, "\000")
		exitStatus, err = strconv.Atoi(parts[0])
		if err != nil {
			panic(err)
		}
	}

	return exitStatus
}
