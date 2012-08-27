package zeusmaster

func StartFileMonitor(tree *ProcessTree, quit chan bool) {
	println("RUNNING FILEMONITOR")

	<- quit
	quit <- true
}
