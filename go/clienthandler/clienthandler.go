package clienthandler

import (
	"errors"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/burke/zeus/go/messages"
	"github.com/burke/zeus/go/processtree"
	slog "github.com/burke/zeus/go/shinylog"
	"github.com/burke/zeus/go/unixsocket"
	"github.com/burke/zeus/go/zerror"
)

func Start(tree *processtree.ProcessTree, done chan bool) chan bool {
	quit := make(chan bool)
	go func() {
		path, _ := filepath.Abs(unixsocket.ZeusSockName())
		addr, err := net.ResolveUnixAddr("unix", path)
		if err != nil {
			zerror.Error("Can't open socket.")
		}
		listener, err := net.ListenUnix("unix", addr)
		if err != nil {
			zerror.ErrorCantCreateListener()
		}

		connections := make(chan *unixsocket.Usock)
		go func() {
			for {
				if conn, err := listener.AcceptUnix(); err != nil {
					zerror.ErrorUnableToAcceptSocketConnection()
					time.Sleep(500 * time.Millisecond)
				} else {
					connections <- unixsocket.New(conn)
				}
			}
		}()

		for {
			select {
			case <-quit:
				listener.Close()
				done <- true
				return
			case conn := <-connections:
				go handleClientConnection(tree, conn)
			}
		}
	}()
	return quit
}

// see docs/client_master_handshake.md
func handleClientConnection(tree *processtree.ProcessTree, usock *unixsocket.Usock) {
	defer usock.Close()
	// we have established first contact to the client.

	command, clientPid, argCount, argFD, err := receiveCommandArgumentsAndPid(usock, nil)

	clientFile, err := receiveTTY(usock, err)
	defer clientFile.Close()

	commandUsock, command, err := bootNewCommand(tree, command, err)
	if err != nil {
		// If a client connects while the command is just
		// booting up, it actually makes it here - still
		// expects a backtrace, of course.
		writeStacktrace(usock, err, clientFile)
		return
	}
	defer commandUsock.Close()

	err = sendClientPidAndArgumentsToCommand(commandUsock, clientPid, argCount, argFD, err)

	err = sendTTYToCommand(commandUsock, clientFile, err)

	cmdPid, err := receivePidFromCommand(commandUsock, err)

	err = sendCommandPidToClient(usock, cmdPid, err)

	exitStatus, err := receiveExitStatus(commandUsock, err)

	err = sendExitStatus(usock, exitStatus, err)

	if err != nil {
		slog.Error(err)
	}
	// Done! Hooray!
}

func writeStacktrace(usock *unixsocket.Usock, err error, clientFile *os.File) {
	// Fake process ID / output / error codes:
	// Write a fake pid (step 6)
	usock.WriteMessage("0")
	// Write the error message to the terminal
	clientFile.Write([]byte(err.Error()))
	// Write a non-positive exit code to the client
	usock.WriteMessage("1")
}

func receiveFileFromFD(usock *unixsocket.Usock) (*os.File, error) {
	clientFd, err := usock.ReadFD()
	if err != nil {
		return nil, errors.New("Expected FD, none received!")
	}
	fileName := strconv.Itoa(rand.Int())
	return os.NewFile(uintptr(clientFd), fileName), nil
}

func receiveCommandArgumentsAndPid(usock *unixsocket.Usock, err error) (string, int, int, int, error) {
	if err != nil {
		return "", -1, -1, -1, err
	}

	msg, err := usock.ReadMessage()
	if err != nil {
		return "", -1, -1, -1, err
	}

	argCount, clientPid, command, err := messages.ParseClientCommandRequestMessage(msg)
	if err != nil {
		return "", -1, -1, -1, err
	}

	argFD, err := usock.ReadFD()
	return command, clientPid, argCount, argFD, err
}

func receiveTTY(usock *unixsocket.Usock, err error) (*os.File, error) {
	if err != nil {
		return nil, err
	}
	return receiveFileFromFD(usock)
}

func sendClientPidAndArgumentsToCommand(commandUsock *unixsocket.Usock, clientPid int, argCount int, argFD int, err error) error {
	if err != nil {
		return err
	}

	msg := messages.CreatePidAndArgumentsMessage(clientPid, argCount)
	_, err = commandUsock.WriteMessage(msg)
	if err != nil {
		return err
	}

	return commandUsock.WriteFD(argFD)
}

func receiveExitStatus(commandUsock *unixsocket.Usock, err error) (string, error) {
	if err != nil {
		return "", err
	}

	return commandUsock.ReadMessage()
}

func sendExitStatus(usock *unixsocket.Usock, exitStatus string, err error) error {
	if err != nil {
		return err
	}

	_, err = usock.WriteMessage(exitStatus)
	return err
}

func receivePidFromCommand(commandUsock *unixsocket.Usock, err error) (int, error) {
	if err != nil {
		return -1, err
	}

	msg, err := commandUsock.ReadMessage()
	if err != nil {
		return -1, err
	}
	intPid, _, _, _ := messages.ParsePidMessage(msg)

	return intPid, err
}

func sendCommandPidToClient(usock *unixsocket.Usock, pid int, err error) error {
	if err != nil {
		return err
	}

	strPid := strconv.Itoa(pid)
	_, err = usock.WriteMessage(strPid)

	return err
}

func bootNewCommand(tree *processtree.ProcessTree, command string, err error) (*unixsocket.Usock, string, error) {
	if err != nil {
		return nil, "", err
	}

	sock, name, err := tree.BootCommand(command)
	if err != nil {
		return nil, "", err
	}

	return sock, name, nil
}

func sendTTYToCommand(commandUsock *unixsocket.Usock, clientFile *os.File, err error) error {
	if err != nil {
		return err
	}

	return commandUsock.WriteFD(int(clientFile.Fd()))
}
