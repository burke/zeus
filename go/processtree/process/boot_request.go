package process

import (
	"os"

	"github.com/burke/zeus/go/messages"
	"github.com/burke/zeus/go/unixsocket"
)

type bootResponse struct {
	Socket *unixsocket.Usock
	Err    error
}

type BootRequest struct {
	response chan bootResponse
	message  string
}

func (br BootRequest) respond(socket *unixsocket.Usock) {
	br.response <- bootResponse{Socket: socket}
	// Only allow responding once
	close(br.response)
}

func (br BootRequest) Error(err error) {
	br.response <- bootResponse{Err: err}
	// Only allow responding once
	close(br.response)
}

func (br BootRequest) String() string {
	return br.message
}

type NodeRequest struct {
	BootRequest
}

func NewNodeRequest(name string) NodeRequest {
	return NodeRequest{
		BootRequest{
			response: make(chan bootResponse, 1),
			message:  messages.CreateSpawnSlaveMessage(name),
		},
	}
}

func (nr NodeRequest) Await() (NodeProcess, error) {
	resp := <-nr.response
	if resp.Err != nil {
		return nil, resp.Err
	}

	proc, err := monitorProcess(resp.Socket)
	if err != nil {
		return nil, err
	}

	return proc, nil
}

type CommandClient struct {
	Pid, ArgCount, ArgFD int
	File                 *os.File
}

type CommandRequest struct {
	BootRequest
	CommandClient
}

func NewCommandRequest(
	name string,
	client CommandClient,

) CommandRequest {
	return CommandRequest{
		BootRequest{
			response: make(chan bootResponse, 1),
			message:  messages.CreateSpawnCommandMessage(name),
		},
		client,
	}
}

func (cr CommandRequest) Await() (CommandProcess, error) {
	resp := <-cr.response
	if resp.Err != nil {
		return nil, resp.Err
	}

	proc := commandProcess{
		socket: resp.Socket,
		client: cr.CommandClient,
		errors: make(chan error),
		ready:  make(chan int),
		wait:   make(chan string),
	}

	go func() {
		if err := proc.run(); err != nil {
			proc.errors <- err
		}
	}()

	return &proc, nil
}
