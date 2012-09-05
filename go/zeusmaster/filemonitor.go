package zeusmaster

import (
	"io"
	"path"
	"strings"
	"os"
	"os/exec"

	slog "github.com/burke/zeus/go/shinylog"
)

var filesToWatch chan string
var filesChanged chan string

var watcherIn io.WriteCloser
var watcherOut io.ReadCloser
var watcherErr io.ReadCloser

var allWatchedFiles map[string]bool

func StartFileMonitor(tree *ProcessTree, quit chan bool) {
	// this is obscenely large, just because as long as we start
	// watching the files eventually, it's more of a priority to 
	// get the slaves booted as quickly as possible.
	filesToWatch = make(chan string, 8192)
	filesChanged = make(chan string, 256)
	allWatchedFiles = make(map[string]bool)

	startWrapper()

	for {
		select {
		case <- quit:
			quit <- true
			return
		case path := <- filesToWatch:
			go handleLoadedFileNotification(path)
		case path := <- filesChanged:
			go handleChangedFileNotification(tree, path)
		}
	}
}

func startWrapper() {
	executable := path.Join(path.Dir(os.Args[0]), "fsevents-wrapper")
	cmd := exec.Command(executable)
	var err error
	if watcherIn, err = cmd.StdinPipe(); err != nil {
		panic(err)
	}
	if watcherOut, err = cmd.StdoutPipe(); err != nil {
		panic(err)
	}
	if watcherErr, err = cmd.StderrPipe() ; err != nil {
		panic(err)
	}

	cmd.Start()

	go func() {
		buf := make([]byte, 2048)
		for {
			n, err := watcherOut.Read(buf)
			if err != nil {
				errorFailedReadFromWatcher(err)
			}
			message := strings.TrimSpace(string(buf[:n]))
			files := strings.Split(message, "\n")
			for _, file := range files {
				filesChanged <- file
			}
		}
	}()

	go func() {
		err := cmd.Wait()
		ErrorFileMonitorWrapperCrashed(err)
	}()
}

func AddFile(file string) {
	filesToWatch <- file
}

func handleLoadedFileNotification(file string) {
	// a slave loaded a file. It's up to us here to make sure this file is watched.
	if !allWatchedFiles[file] {
		allWatchedFiles[file] = true
		startWatchingFile(file)
	}
}

func handleChangedFileNotification(tree *ProcessTree, file string) {
	slog.Yellow("Dependency change at " + file)
	tree.KillNodesWithFeature(file)
}


func startWatchingFile(file string) {
	watcherIn.Write([]byte(file + "\n"))
}
