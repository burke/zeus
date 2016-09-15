package process

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"golang.org/x/sys/unix"

	"github.com/burke/zeus/go/messages"
	slog "github.com/burke/zeus/go/shinylog"
	"github.com/burke/zeus/go/unixsocket"
)

const (
	forceKillTimeout = time.Second
)

var ErrProcessStopping = errors.New("process is stopping")

type NodeProcess interface {
	Errors() <-chan error
	Ready() <-chan struct{}
	// TODO: Switch all instances of "identifier" to "name"
	Identifier() string
	Pid() int
	ParentPid() int
	Files() <-chan string
	Boot(BootRequest)
	Stop() error
}

type nodeProcess struct {
	errors         chan error
	files          chan string
	identifier     string
	ready          chan struct{}
	stop           chan struct{}
	pid, parentPid int
	socket         *unixsocket.Usock
	requests       chan BootRequest
}

func StartProcess(args []string) (NodeProcess, error) {
	localFile, remoteFile, err := unixsocket.Socketpair(syscall.SOCK_DGRAM)
	if err != nil {
		return nil, err
	}
	defer unix.Close(int(remoteFile.Fd()))
	defer unix.Close(int(localFile.Fd()))

	localSocket, err := unixsocket.NewFromFile(localFile)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(args[0], args[1:]...)

	buf := prefixSuffixSaver{N: 32 << 10}
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	cmd.Env = append(os.Environ(), fmt.Sprintf("ZEUS_MASTER_FD=%d", remoteFile.Fd()))
	cmd.ExtraFiles = []*os.File{remoteFile}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	cmd.Run()

	errCh := make(chan error, 1)
	procCh := make(chan *nodeProcess, 1)

	go func() {
		err := cmd.Wait()
		var msg string
		if err == nil {
			msg = fmt.Sprintf("process exited with zero exit code")
			// Exited unexpectedly, no code
		} else if exitErr, ok := err.(*exec.ExitError); ok {
			msg = fmt.Sprintf("process %d exited with non-zero exit code", exitErr.Pid())
		} else {
			msg = fmt.Sprintf("error waiting for process: %v", err)
		}

		errCh <- fmt.Errorf("%s. Output was:\n%s", msg, buf.Bytes())
	}()

	go func() {
		sock, err := localSocket.ReadSocket()
		if err != nil {
			errCh <- err
			return
		}

		proc, err := monitorProcess(sock)
		if err != nil {
			errCh <- err
			return
		}

		procCh <- proc
	}()

	var proc *nodeProcess
	select {
	case err := <-errCh:
		return nil, fmt.Errorf("error starting process: %v", err)
	case proc = <-procCh:
	}

	go func() {
		select {
		case <-proc.stop:
			// Exited as expected
		case err := <-errCh:
			proc.errors <- err
		}
	}()

	return proc, err
}

func monitorProcess(socket *unixsocket.Usock) (*nodeProcess, error) {
	// TODO: prevent this from blocking when process is killed
	msg, err := socket.ReadMessage()
	if err != nil {
		return nil, err
	}

	pid, parentPid, identifier, err := messages.ParsePidMessage(msg)
	if err != nil {
		return nil, err
	}

	p := nodeProcess{
		errors:     make(chan error),
		files:      make(chan string),
		identifier: identifier,
		pid:        pid,
		parentPid:  parentPid,
		socket:     socket,
		requests:   make(chan BootRequest, 128),
		stop:       make(chan struct{}),
		ready:      make(chan struct{}),
	}

	// Crashing is stop followed by writing an error to a channel?
	go func() {
		if err := p.run(); err != nil {
			p.errors <- err
		}
	}()

	return &p, nil
}

func (p *nodeProcess) Errors() <-chan error {
	return p.errors
}

func (p *nodeProcess) Identifier() string {
	return p.identifier
}

func (p *nodeProcess) Pid() int {
	return p.pid
}

func (p *nodeProcess) ParentPid() int {
	return p.parentPid
}

func (p *nodeProcess) Files() <-chan string {
	return p.files
}

func (p *nodeProcess) Ready() <-chan struct{} {
	return p.ready
}

func (p *nodeProcess) run() error {
	// And the last step before executing its action, the slave sends us a pipe it will later use to
	// send us all the features it's loaded.

	// TODO: prevent this from blocking when process is killed
	featurePipeFd, err := p.socket.ReadFD()
	if err != nil {
		return err
	}
	featurePipe := os.NewFile(uintptr(featurePipeFd), "featurepipe")
	go p.handleFileMessages(featurePipe)

	// The slave will execute its action and respond with a status...
	msg, err := p.socket.ReadMessage()
	if err != nil {
		return err
	}

	msg, err = messages.ParseActionResponseMessage(msg)
	if err != nil {
		return err
	}

	if msg == "OK" {
		close(p.ready)
		return p.handleRequests()
	}
	return errors.New(msg)
}

func (p *nodeProcess) handleFileMessages(featurePipe *os.File) {
	reader := bufio.NewReader(featurePipe)
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				// EOF indicates that the process went away and we should close the socket
				if err := p.socket.Close(); err != nil {
					slog.ErrorString(fmt.Sprintf("error closing socket after file message pipe closed for process %s with pid %d: %v", p.Identifier(), p.Pid(), err))
				}
			} else {
				slog.ErrorString(fmt.Sprintf("error handling file messages for process %s with pid %d: %v", p.Identifier(), p.Pid(), err))

			}
			break
		}

		msg = strings.TrimRight(msg, "\n")
		// We don't want to pass empty strings since they could be confused
		// for a closed channel by the reader.
		if msg != "" {
			p.files <- msg
		}
	}
	close(p.files)
}

func (p *nodeProcess) handleRequests() error {
	for {
		select {
		case <-p.stop:
			return nil
		case req := <-p.requests:
			sock, err := p.sendBootMessage(req.message)

			// If we encountered an error sending the message the
			// pipe could be in an inconsistent state, we can't safely
			// send further messages.
			if err != nil {
				req.Error(err)
				return err
			}

			req.respond(sock)
		}
	}
}

func (p *nodeProcess) sendBootMessage(msg string) (*unixsocket.Usock, error) {
	_, err := p.socket.WriteMessage(msg)
	if err != nil {
		return nil, err
	}

	sock, err := p.socket.ReadSocket()
	if err != nil {
		return nil, err
	}

	return sock, nil
}

// Does not block waiting for boot, unsafe to call in parallel with Stop
func (p *nodeProcess) Boot(req BootRequest) {
	select {
	case <-p.stop:
		req.Error(ErrProcessStopping)
		return
	default:
	}

	select {
	case p.requests <- req:
	default:
		req.Error(errors.New("request queue for process %d is full"))
	}
}

// Blocking, not concurrency-safe
func (p *nodeProcess) Stop() error {
	select {
	case <-p.stop:
		return errors.New("Process has already stopped")
	default:
		close(p.stop)
		close(p.requests)
	}

	if err := p.forceKill(); err != nil {
		return err
	}

	for req := range p.requests {
		req.Error(ErrProcessStopping)
	}

	return nil
}

func (p *nodeProcess) forceKill() error {
	if err := syscall.Kill(p.pid, syscall.SIGTERM); err != nil {
		return fmt.Errorf("Error killing pid %q: %v", p.pid, err)
	}

	exited := make(chan error)
	go func() {
		for {
			if err := syscall.Kill(p.pid, syscall.Signal(0)); err != nil {
				exited <- nil
				return
			}

			// Since the process is not our direct child, we can't use wait
			// and are forced to poll for completion. We know this won't loop
			// forever because the timeout below will SIGKILL the process
			// which guarantees that it'll go away and we'll get an ESRCH.
			time.Sleep(time.Millisecond)
		}
	}()

	select {
	case err := <-exited:
		if err != nil && err != syscall.ESRCH {
			return fmt.Errorf("Error sending signal to pid %q: %v", p.pid, err)
		}
		return nil
	case <-time.After(forceKillTimeout):
		syscall.Kill(p.pid, syscall.SIGKILL)
		return nil
	}
}
