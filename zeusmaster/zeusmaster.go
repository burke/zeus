package zeusmaster

import (
	"syscall"
	"time"

	usock "github.com/burke/zeus/unixsocket"
)

func Run() {
	var tree *ProcessTree = BuildProcessTree()

	localMasterSocket, remoteMasterSocket, err := usock.Socketpair(syscall.SOCK_DGRAM)
	if err != nil {
		panic(err)
	}

	localMasterUNIXSocket, err := usock.MakeUnixSocket(localMasterSocket)
	if err != nil {
		panic(err)
	}

	go StartSlaveMonitor(tree, localMasterUNIXSocket, remoteMasterSocket)
	go StartClientHandler(tree)
	go StartFileMonitor(tree)

	for {
		// is there a better way to sleep forever?
		time.Sleep(1000 * time.Second)
	}
}


