package zeusmaster

import "fmt"

type fileNotification struct {
	identifier string
	file string
}

var files chan *fileNotification

func StartFileMonitor(tree *ProcessTree, quit chan bool) {
	// this is obscenely large, just because as long as we start
	// watching the files eventually, it's more of a priority to 
	// get the slaves booted as quickly as possible.
	files = make(chan *fileNotification, 5000)

	for {
		select {
		case <- quit:
			quit <- true
			return
		case notif := <- files:
			go handleFileNotification(notif.identifier, notif.file)
		}
	}
}

func AddFile(identifier, file string) {
	files <- &fileNotification{identifier, file}
}

func handleFileNotification(identifier, file string) {
	fmt.Println(identifier, file)
}
