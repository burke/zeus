package zeusmaster

import (
	"fmt"
	"net"
	"path/filepath"
	"os"
)

const zeusSockName string = ".zeus.sock"

func StartClientHandler(tree *ProcessTree) {
	path, _ := filepath.Abs(zeusSockName)
	addr, err := net.ResolveUnixAddr("unix", path)
	if err != nil {
		panic("Can't open socket.")
	}
	listener, err := net.ListenUnix("unix", addr)
	if err != nil {
		panic("Can't create listener")
	}
	defer listener.Close()
	defer removeSock(path)
	for {
		conn, err := listener.AcceptUnix()
		if err != nil {
			panic("Unable to accept Socket connection")
		}
		go handleClientConnection(conn)
	}
}

func removeSock(path string) {
	err := os.Remove(path)
	if err != nil { 
		fmt.Println(err)
	} else {
		fmt.Println("REMOVED ", path)
	}
}

func handleClientConnection(conn net.UnixConn) {

	// we have established first contact to the client.

	// we first read the command and arguments specified from the connection.

	// Now we read the terminal IO socket to use for raw IO

	// Now we read a socket the client wants to use for exit status
	// (can we just use the parameter conn?)

}




}
