package process

import (
	"github.com/burke/zeus/go/messages"
	"github.com/burke/zeus/go/unixsocket"
)

type bootResponse struct {
	Socket *unixsocket.Usock
	Err    error
}

// BootRequest represents a request for a process to fork a new child.
type BootRequest struct {
	response chan bootResponse
	message  string
}

func (br BootRequest) respond(socket *unixsocket.Usock) {
	br.response <- bootResponse{Socket: socket}
	// Only allow responding once
	close(br.response)
}

// Error reports an error servicing the boot request.
func (br BootRequest) Error(err error) {
	br.response <- bootResponse{Err: err}
	// Only allow responding once
	close(br.response)
}

func (br BootRequest) String() string {
	return br.message
}

// NodeRequest represents a request to boot a NodeProcess that
// can be used to boot further child processes.
type NodeRequest struct {
	BootRequest
}

// NewNodeRequest creates a new request for boot a NodeProcess with
// the specified action name.
func NewNodeRequest(name string) NodeRequest {
	return NodeRequest{
		BootRequest{
			response: make(chan bootResponse, 1),
			message:  messages.CreateSpawnSlaveMessage(name),
		},
	}
}

// Await blocks until the request has been serviced and either
// returns a new NodeProcess or an error.
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

// CommandRequest represents a request to boot a command that will
// connect to a TTY.
type CommandRequest struct {
	BootRequest
}

// NewCommandRequest creates a new request to boot a command with
// the specified action name.
func NewCommandRequest(
	name string,

) CommandRequest {
	return CommandRequest{
		BootRequest{
			response: make(chan bootResponse, 1),
			message:  messages.CreateSpawnCommandMessage(name),
		},
	}
}

// Await blocks until the request has been servied and either
// returns a socket for communicating with the process or an error.
// In future versions this interface will be changed to return a CommandProcess
// type that wraps interaction with the socket.
func (cr CommandRequest) Await() (*unixsocket.Usock, error) {
	resp := <-cr.response
	if resp.Err != nil {
		return nil, resp.Err
	}
	return resp.Socket, nil
}
