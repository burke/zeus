package zeusmaster

func StartFileMonitor(tree *ProcessTree, quit chan bool) {

	for {
		select {
		case <- quit:
			quit <- true
			return
		}
	}
}
