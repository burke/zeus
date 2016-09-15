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
	"github.com/burke/zeus/go/processtree/process"
	slog "github.com/burke/zeus/go/shinylog"
	"github.com/burke/zeus/go/unixsocket"
	"github.com/burke/zeus/go/zerror"
)

func Start(tree processtree.ProcessTree, done chan bool) chan bool {
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
				go func() {
					defer conn.Close()
					if err := handleClientConnection(tree, conn); err != nil {
						slog.Error(err)
					}
				}()
			}
		}
	}()
	return quit
}

// see docs/client_master_handshake.md
func handleClientConnection(tree processtree.ProcessTree, usock *unixsocket.Usock) error {
	// we have established first contact to the client.
	command, clientPid, argCount, argFD, err := receiveCommandArgumentsAndPid(usock, nil)
	if err != nil {
		return err
	}

	clientFile, err := receiveFileFromFD(usock)
	if err != nil {
		return err
	}
	defer clientFile.Close()

	proc, err := tree.BootCommand(command, process.CommandClient{
		Pid:      clientPid,
		ArgCount: argCount,
		ArgFD:    argFD,
		File:     clientFile,
	})
	if err != nil {
		return err
	}
	defer proc.Stop()

	select {
	case pid := <-proc.Ready():
		if err := sendCommandPidToClient(usock, pid, err); err != nil {
			return err
		}
	case err := <-proc.Errors():
		writeStacktrace(usock, clientFile, err)
		return nil
	}

	select {
	case exitStatus := <-proc.Wait():
		if err := sendExitStatus(usock, exitStatus, err); err != nil {
			return err
		}
	case err := <-proc.Errors():
		return err
	}

	return nil
}

func writeStacktrace(usock *unixsocket.Usock, clientFile *os.File, err error) {
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

func sendExitStatus(usock *unixsocket.Usock, exitStatus string, err error) error {
	if err != nil {
		return err
	}

	_, err = usock.WriteMessage(exitStatus)
	return err
}

func sendCommandPidToClient(usock *unixsocket.Usock, pid int, err error) error {
	if err != nil {
		return err
	}

	strPid := strconv.Itoa(pid)
	_, err = usock.WriteMessage(strPid)

	return err
}
