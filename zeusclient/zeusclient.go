package zeusclient

import (
	"os"
	"os/signal"
	"net"
	"strings"
	"strconv"
	"syscall"

	"github.com/kr/pty"
	"github.com/burke/ttyutils"
	usock "github.com/burke/zeus/unixsocket"
)

const (
	zeusSockName = ".zeus.sock"
)

func Run() {
	master, slave, err := pty.Open()
	if err != nil {
		panic(err)
	}
	defer master.Close()
	defer slave.Close()

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

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGWINCH)
	go func() {
		for _ = range c {
			ttyutils.MirrorWinsize(os.Stdout, master)
			syscall.Kill(pid, syscall.SIGWINCH)
		}
	}()

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

