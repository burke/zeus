package process

import (
	"fmt"

	"github.com/burke/zeus/go/messages"
	"github.com/burke/zeus/go/unixsocket"
)

type CommandProcess interface {
	Ready() <-chan int
	Errors() <-chan error
	Wait() <-chan string
	Stop() error
}

type commandProcess struct {
	socket *unixsocket.Usock
	errors chan error
	ready  chan int
	wait   chan string
	client CommandClient
}

func (p *commandProcess) Ready() <-chan int {
	return p.ready
}

func (p *commandProcess) Errors() <-chan error {
	return p.errors
}

func (p *commandProcess) Wait() <-chan string {
	return p.wait
}

func (p *commandProcess) Stop() error {
	// TODO: Should we do something more graceful here?
	return p.socket.Close()
}

func (p *commandProcess) run() error {
	msg := messages.CreatePidAndArgumentsMessage(p.client.Pid, p.client.ArgCount)
	if _, err := p.socket.WriteMessage(msg); err != nil {
		return fmt.Errorf("error sending pid and argument count: %v", err)
	}

	if err := p.socket.WriteFD(p.client.ArgFD); err != nil {
		return fmt.Errorf("error sending arg file descriptor: %v", err)
	}

	if err := p.socket.WriteFD(int(p.client.File.Fd())); err != nil {
		return fmt.Errorf("error sending client TTY: %v", err)
	}

	msg, err := p.socket.ReadMessage()
	if err != nil {
		return fmt.Errorf("error receiving command pid: %v", err)
	}
	intPid, _, _, err := messages.ParsePidMessage(msg)
	if err != nil {
		return err
	}
	p.ready <- intPid

	exitStatus, err := p.socket.ReadMessage()
	if err != nil {
		return fmt.Errorf("error receiving command exit status: %v", err)
	}
	p.wait <- exitStatus

	return nil
}
