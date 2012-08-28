package zeusmaster

import (
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"path/filepath"
	"os"

	usock "github.com/burke/zeus/unixsocket"
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
			go handleClientConnection(tree, conn)
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

// see docs/client_master_handshake.md
func handleClientConnection(tree *ProcessTree, conn *net.UnixConn) {
	defer conn.Close()

	// we have established first contact to the client.

	// we first read the command and arguments specified from the connection. (step 1)
	msg, _, err := usock.ReadFromUnixSocket(conn)
	if err != nil {
		fmt.Println(err)
		return
	}
	command, arguments, err := ParseClientCommandRequestMessage(msg)
	if err != nil {
		fmt.Println(err)
		return
	}

	commandNode := tree.FindCommandByName(command)
	if commandNode == nil {
		fmt.Println("ERROR: Node not found!: ", command)
		return
	}
	slaveNode := commandNode.Parent

	// Now we read the terminal IO socket to use for raw IO (step 2)
	clientFd, err := usock.ReadFileDescriptorFromUnixSocket(conn)
	if err != nil {
		fmt.Println("Expected FD, none received!")
		return
	}
	fileName := strconv.Itoa(rand.Int())
	clientFile := usock.FdToFile(clientFd, fileName)
	defer clientFile.Close()

	// We now need to fork a new command process.
	// For now, we naively assume it's running...

	// boot a command process and establish a socket connection to it.
	slaveNode.WaitUntilBooted()

	if slaveNode.Error != "" {
		// we can skip steps 3-5 as they deal with the command process we're not spawning.
		// Write a fake pid (step 6)
		conn.Write([]byte("0\n"))
		// Write the error message to the terminal
		clientFile.Write([]byte(slaveNode.Error))
		// Skip step 7, and write an exit code to the client (step 8)
		conn.Write([]byte("1\n"))
		return
	}

	slaveNode.mu.Lock()
	slaveNode.Socket.Write([]byte("C:console"))

	commandFd, err := usock.ReadFileDescriptorFromUnixSocket(slaveNode.Socket)
	slaveNode.mu.Unlock()
	if err != nil {
		fmt.Println("Couldn't start command process!", err)
	}
	fileName = strconv.Itoa(rand.Int())
	commandFile := usock.FdToFile(commandFd, fileName)
	defer commandFile.Close()

	// Send the arguments to the command process (step 3)
	commandFile.Write([]byte(arguments)) // TODO: What if they're too long?

	commandSocket, err := usock.MakeUnixSocket(commandFile)
	if err != nil {
		fmt.Println("MakeUnixSocket", err)
	}
	defer commandSocket.Close()

	// Send the client terminal connection to the command process (step 4)
	usock.SendFdOverUnixSocket(commandSocket, clientFd)

	// Receive the pid from the command process (step 5)
	msg, _, err = usock.ReadFromUnixSocket(commandSocket)
	if err != nil {
		println(err)
	}
	intPid, _, _ := ParsePidMessage(msg)

	// Send the pid to the client process (step 6)
	strPid := strconv.Itoa(intPid)
	conn.Write([]byte(strPid + "\n"))

	// Receive the exit status from the command (step 7)
	msg, _, err = usock.ReadFromUnixSocket(commandSocket)
	if err != nil {
		println(err)
	}

	// Forward the exit status to the Client (step 8)
	conn.Write([]byte(msg))

	// Done! Hooray!

}
