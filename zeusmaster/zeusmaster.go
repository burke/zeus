package zeusmaster

import (
	"syscall"
	"time"
	// "os"
	// "os/signal"

	usock "github.com/burke/zeus/unixsocket"
)

func Run() {
	var tree *ProcessTree = BuildProcessTree()

	// graceful shutdown...
	// c := make(chan os.Signal, 1)
	// signal.Notify(c, os.Interrupt)
	// go function(){
	// 	for sig := range c {
	// 		// sig is a ^C, handle it
	// 	}
	// }()

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


