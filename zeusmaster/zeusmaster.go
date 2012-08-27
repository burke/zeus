package zeusmaster

import (
	"syscall"
	"os"
	"os/signal"

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

	quit1 := make(chan bool, 1)
	quit2 := make(chan bool, 1)
	quit3 := make(chan bool, 1)

	StartSlaveMonitor(tree, localMasterUNIXSocket, remoteMasterSocket, quit1)
	StartClientHandler(tree, quit2)
	StartFileMonitor(tree, quit3)

	quit := make(chan bool, 1)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func(){
		for _ = range c {
			quit1 <- true
			quit2 <- true
			quit3 <- true
			<- quit1
			<- quit2
			<- quit3
			quit <- true
		}
	}()

	<- quit
}


