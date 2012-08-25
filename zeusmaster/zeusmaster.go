package zeusmaster

import (
	"time"
)


func Run() {
	var tree *ProcessTree = BuildProcessTree()
	go StartSlaveMonitor(tree)
	go StartClientHandler(tree)
	go StartFileMonitor(tree)

	time.Sleep(500 * time.Millisecond)
}


