package zeusmaster

func StartFileMonitor(tree *ProcessTree, quit chan bool) {
	println("RUNNING FILEMONITOR")

	for {
		select {
		case <- quit:
			quit <- true
			return
		}
	}
}
