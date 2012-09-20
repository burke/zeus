package zeusmaster

import (
	"io"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"

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
		case <-quit:
			quit <- true
			return
		case path := <-filesToWatch:
			go handleLoadedFileNotification(path)
		case path := <-filesChanged:
			go handleChangedFileNotification(tree, path)
		}
	}
}

func executablePath() string {
	switch runtime.GOOS {
	case "darwin":
		return path.Join(path.Dir(os.Args[0]), "fsevents-wrapper")
	case "linux":
		gemRoot := path.Dir(path.Dir(os.Args[0]))
		return path.Join(gemRoot, "ext/inotify-wrapper/inotify-wrapper")
	}
	Error("Unsupported OS")
	return ""
}

func startWrapper() {
	cmd := exec.Command(executablePath())
	var err error
	if watcherIn, err = cmd.StdinPipe(); err != nil {
		Error(err.Error())
	}
	if watcherOut, err = cmd.StdoutPipe(); err != nil {
		Error(err.Error())
	}
	if watcherErr, err = cmd.StderrPipe(); err != nil {
		Error(err.Error())
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
	tree.RestartNodesWithFeature(file)
}

func startWatchingFile(file string) {
	_, err := watcherIn.Write([]byte(file + "\n"))
	if err != nil {
		slog.Error(err)
	}
}
