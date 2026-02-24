package zeusclient

import (
	"io"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/burke/ttyutils"
	"github.com/burke/zeus/go/messages"
	slog "github.com/burke/zeus/go/shinylog"
	"github.com/burke/zeus/go/unixsocket"
	"github.com/burke/zeus/go/zerror"
	"github.com/kr/pty"
)

const (
	sigInt  = 3
	sigQuit = 28
	sigTstp = 26
)

// man signal | grep 'terminate process' | awk '{print $2}' | xargs -I '{}' echo -n "syscall.{}, "
var terminatingSignals = []os.Signal{syscall.SIGHUP, syscall.SIGINT, syscall.SIGKILL, syscall.SIGPIPE, syscall.SIGALRM, syscall.SIGTERM, syscall.SIGXCPU, syscall.SIGXFSZ, syscall.SIGVTALRM, syscall.SIGPROF, syscall.SIGUSR1, syscall.SIGUSR2}

// Run runs the main Zeus command.
func Run(args []string, input io.Reader, output *os.File, stderr *os.File, ttyMode string) int {
	if os.Getenv("RAILS_ENV") != "" {
		println("Warning: Specifying a Rails environment via RAILS_ENV has no effect for commands run with zeus.")
		println("As a safety precaution to protect you from nuking your development database,")
		println("Zeus will now cowardly refuse to proceed. Please unset RAILS_ENV and try again.")
		return 1
	}

	// setup stdout
	localStdout, remoteStdout, outputIsTerminal, err := socketsForOutput(output, ttyMode)
	if err != nil {
		slog.ErrorString(err.Error() + "\r")
		return 1
	}
	defer localStdout.Close()

	// setup stderr
	localStderr, remoteStderr, _, err := socketsForOutput(stderr, ttyMode)
	if err != nil {
		slog.ErrorString(err.Error() + "\r")
		return 1
	}
	defer localStderr.Close()

	addr, err := net.ResolveUnixAddr("unixgram", unixsocket.ZeusSockName())
	if err != nil {
		slog.ErrorString(err.Error() + "\r")
		return 1
	}

	conn, err := net.DialUnix("unix", nil, addr)
	if err != nil {
		zerror.ErrorCantConnectToMaster()
		return 1
	}
	usock := unixsocket.New(conn)

	msg := messages.CreateCommandAndArgumentsMessage(args, os.Getpid())
	usock.WriteMessage(msg)
	err = sendCommandLineArguments(usock, args)
	if err != nil {
		slog.ErrorString(err.Error() + "\r")
		return 1
	}

	usock.WriteFD(int(remoteStdout.Fd()))
	usock.WriteFD(int(remoteStderr.Fd()))

	// Close client's copies of the remote ends now that they've been sent
	// to the server. Otherwise the local ends never get EOF when the
	// command exits, because these fds keep the socket pairs alive.
	remoteStdout.Close()
	remoteStderr.Close()

	msg, err = usock.ReadMessage()
	if err != nil {
		slog.ErrorString(err.Error() + "\r")
		return 1
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
		slog.ErrorString(err.Error() + "\r")
		return 1
	}

	if outputIsTerminal {
		c := make(chan os.Signal, 1)
		handledSignals := append(append(terminatingSignals, syscall.SIGWINCH), syscall.SIGCONT)
		signal.Notify(c, handledSignals...)
		go func() {
			for sig := range c {
				if sig == syscall.SIGCONT {
					syscall.Kill(commandPid, syscall.SIGCONT)
				} else if sig == syscall.SIGWINCH {
					ttyutils.MirrorWinsize(output, localStdout)
					syscall.Kill(commandPid, syscall.SIGWINCH)
				} else { // member of terminatingSignals
					print("\r")
					syscall.Kill(commandPid, sig.(syscall.Signal))
					os.Exit(1)
				}
			}
		}()
	}

	exitStatus := -1
	if len(parts) > 2 {
		exitStatus, err = strconv.Atoi(parts[0])
		if err != nil {
			slog.ErrorString(err.Error() + "\r")
			return 1
		}
	}

	oldTerminalStateStdout, err := forwardOutput(localStdout, output)
	if oldTerminalStateStdout != nil {
		defer ttyutils.RestoreTerminalState(output.Fd(), oldTerminalStateStdout)
	}
	if err != nil {
		slog.ErrorString(err.Error() + "\r")
		return 1
	}

	oldTerminalStateStderr, err := forwardOutput(localStderr, stderr)
	if oldTerminalStateStderr != nil {
		defer ttyutils.RestoreTerminalState(stderr.Fd(), oldTerminalStateStderr)
	}
	if err != nil {
		slog.ErrorString(err.Error() + "\r")
		return 1
	}

	go func() {
		buf := make([]byte, 8192)
		for {
			n, err := input.Read(buf)
			if err != nil {
				// Signal EOF to the remote process. For socket pairs,
				// shut down the write direction so the command's stdin
				// gets EOF instead of blocking forever (e.g. when
				// stdin is /dev/null from an AI tool).
				if !outputIsTerminal {
					syscall.Shutdown(int(localStdout.Fd()), syscall.SHUT_WR)
				}
				break
			}
			if outputIsTerminal {
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
			}
			localStdout.Write(buf[:n])
		}
	}()

	if exitStatus == -1 {
		msg, err = usock.ReadMessage()
		if err != nil {
			slog.ErrorString(err.Error() + "\r")
			return 1
		}
		parts := strings.Split(msg, "\000")
		exitStatus, err = strconv.Atoi(parts[0])
		if err != nil {
			slog.ErrorString(err.Error() + "\r")
			return 1
		}
	}

	// Command has exited and reported its status. Clear commandPid so
	// the defer doesn't SIGKILL it — the process may still be flushing
	// buffered stdout (especially with socket pairs where Ruby uses
	// full buffering). Killing it here loses output.
	commandPid = 0

	return exitStatus
}

func sendCommandLineArguments(usock *unixsocket.Usock, args []string) error {
	master, slave, err := unixsocket.Socketpair(syscall.SOCK_STREAM)
	if err != nil {
		return err
	}
	usock.WriteFD(int(slave.Fd()))
	if err != nil {
		return err
	}
	slave.Close()

	go func() {
		defer master.Close()
		argAsBytes := []byte{}
		for _, arg := range args[1:] {
			argAsBytes = append(argAsBytes, []byte(arg)...)
			argAsBytes = append(argAsBytes, byte(0))
		}
		_, err = master.Write(argAsBytes)
		if err != nil {
			slog.ErrorString("Could not send arguments across: " +
				err.Error() + "\r")
			return
		}
	}()

	return nil
}

func socketsForOutput(out *os.File, ttyMode string) (local, remote *os.File, outIsTerminal bool, err error) {
	switch ttyMode {
	case "force":
		outIsTerminal = true
	case "none":
		outIsTerminal = false
	default:
		outIsTerminal = ttyutils.IsTerminal(out.Fd())
	}

	if outIsTerminal {
		local, remote, err = pty.Open()
	} else {
		local, remote, err = unixsocket.Socketpair(syscall.SOCK_STREAM)
	}

	return
}

func forwardOutput(from, to *os.File) (oldTerminalState *ttyutils.Termios, err error) {
	if ttyutils.IsTerminal(to.Fd()) {
		ttyutils.MirrorWinsize(to, from)
	}

	if ttyutils.IsTerminal(to.Fd()) {
		oldTerminalState, err = ttyutils.MakeTerminalRaw(to.Fd())
		if err != nil {
			return
		}
	}

	go func() {
		for {
			buf := make([]byte, 1024)
			n, err := from.Read(buf)

			if err == nil || (err == io.EOF && n > 0) {
				to.Write(buf[:n])
			} else {
				break
			}
		}
	}()

	return
}
