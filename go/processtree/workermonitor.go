package processtree

import (
	"math/rand"
	"os"
	"strconv"
	"syscall"

	"github.com/burke/zeus/go/messages"
	slog "github.com/burke/zeus/go/shinylog"
	"github.com/burke/zeus/go/unixsocket"
)

type WorkerMonitor struct {
	tree                  *ProcessTree
	remoteCoordinatorFile *os.File
}

func Error(err string) {
	// TODO
	println(err)
}

func StartWorkerMonitor(tree *ProcessTree, fileChanges <-chan []string, done chan bool) chan bool {
	quit := make(chan bool)
	go func() {
		localCoordinatorFile, remoteCoordinatorFile, err := unixsocket.Socketpair(syscall.SOCK_DGRAM)
		if err != nil {
			Error("Couldn't create socketpair")
		}

		monitor := &WorkerMonitor{tree, remoteCoordinatorFile}
		defer monitor.cleanupChildren()

		localCoordinatorSocket, err := unixsocket.NewFromFile(localCoordinatorFile)
		if err != nil {
			Error("Couldn't Open UNIXSocket")
		}

		// We just want this unix socket to be a channel so we can select on it...
		registeringFds := make(chan int, 3)
		go func() {
			for {
				if fd, err := localCoordinatorSocket.ReadFD(); err != nil {
					slog.Error(err)
				} else {
					registeringFds <- fd
				}
			}
		}()

		for _, worker := range monitor.tree.WorkersByName {
			go worker.Run(monitor)
		}

		for {
			select {
			case <-quit:
				done <- true
				return
			case fd := <-registeringFds:
				go monitor.workerDidBeginRegistration(fd)
			case files := <-fileChanges:
				if len(files) > 0 {
					tree.RestartNodesWithFeatures(files)
				}
			}
		}
	}()
	return quit
}

func (mon *WorkerMonitor) cleanupChildren() {
	for _, worker := range mon.tree.WorkersByName {
		worker.ForceKill()
	}
}

func (mon *WorkerMonitor) workerDidBeginRegistration(fd int) {
	// Having just started the process, we expect an IO, which we convert to a UNIX domain socket
	fileName := strconv.Itoa(rand.Int())
	workerFile := os.NewFile(uintptr(fd), fileName)
	workerUsock, err := unixsocket.NewFromFile(workerFile)
	if err != nil {
		slog.Error(err)
	}

	// We now expect the worker to use this fd they send us to send a Pid&Identifier Message
	msg, err := workerUsock.ReadMessage()
	if err != nil {
		slog.Error(err)
	}
	pid, parentPid, identifier, err := messages.ParsePidMessage(msg)

	// And the last step before executing its action, the worker sends us a pipe it will later use to
	// send us all the features it's loaded.
	featurePipeFd, err := workerUsock.ReadFD()
	if err != nil {
		slog.Error(err)
	}

	workerNode := mon.tree.FindWorkerByName(identifier)
	if workerNode == nil {
		Error("workermonitor.go:workerDidBeginRegistration:Unknown identifier:" + identifier)
	}

	workerNode.WorkerWasInitialized(pid, parentPid, workerUsock, featurePipeFd)
}
