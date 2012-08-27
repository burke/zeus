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

	quit1 := make(chan bool)
	quit2 := make(chan bool)
	quit3 := make(chan bool)

	go StartSlaveMonitor(tree, localMasterUNIXSocket, remoteMasterSocket, quit1)
	go StartClientHandler(tree, quit2)
	go StartFileMonitor(tree, quit3)

	quit := make(chan bool, 1)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func(){
		// FIXME: Unprecedented levels of jank, right here.
		for _ = range c {
			go func() {
				quit1 <- true
				<- quit1
				quit <- true
			}()
			go func() {
				quit2 <- true
				<- quit2
				quit <- true
			}()
			go func() {
				quit3 <- true
				<- quit3
				quit <- true
			}()
		}
	}()

	<- quit
	<- quit
	<- quit
}


