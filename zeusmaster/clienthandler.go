package zeusmaster

import (
	"fmt"
	"net"
	"path/filepath"
	"os"
)

const zeusSockName string = ".zeus.sock"

func StartClientHandler(tree *ProcessTree, quit chan bool) {
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

	connections := make(chan *net.UnixConn)
	go func() {
		for {
			if conn, err := listener.AcceptUnix() ; err != nil {
				fmt.Println("Unable to accept Socket connection")
			} else {
				connections <- conn
			}
		}
	}()

	for {
		select {
		case <- quit:
			quit <- true
			return
		case conn := <- connections:
			go handleClientConnection(conn)
		}
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

func handleClientConnection(conn *net.UnixConn) {

	// we have established first contact to the client.

	// we first read the command and arguments specified from the connection.

	// Now we read the terminal IO socket to use for raw IO

	// Now we read a socket the client wants to use for exit status
	// (can we just use the parameter conn?)

}
