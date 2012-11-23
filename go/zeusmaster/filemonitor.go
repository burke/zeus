package zeusmaster

import (
	"io"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
	"sync"

	slog "github.com/burke/zeus/go/shinylog"
)

var filesToWatch chan string
var filesChanged chan string

var watcherIn io.WriteCloser
var watcherOut io.ReadCloser
var watcherErr io.ReadCloser

var fileMutex sync.Mutex

var allWatchedFiles map[string]bool

func StartFileMonitor(tree *ProcessTree) chan bool {
	quit := make(chan bool)
	go func() {
		// this is obscenely large, just because as long as we start
		// watching the files eventually, it's more of a priority to 
		// get the slaves booted as quickly as possible.
		filesToWatch = make(chan string, 8192)
		filesChanged = make(chan string, 256)
		allWatchedFiles = make(map[string]bool)

		cmd := startWrapper()

		for {
			select {
			case <-quit:
				cmd.Process.Kill()
				quit <- true
				return
			case path := <-filesToWatch:
				go handleLoadedFileNotification(path)
			case path := <-filesChanged:
				go handleChangedFileNotification(tree, path)
			}
		}
	}()
	return quit
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

func startWrapper() *exec.Cmd {
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

	return cmd
}

func AddFile(file string) {
	filesToWatch <- file
}

func handleLoadedFileNotification(file string) {
	fileMutex.Lock()
	// a slave loaded a file. It's up to us here to make sure this file is watched.
	if !allWatchedFiles[file] {
		allWatchedFiles[file] = true
		startWatchingFile(file)
	}
	fileMutex.Unlock()
}

func handleChangedFileNotification(tree *ProcessTree, file string) {
	// slog.Magenta("[filechange] " + file)
	fileMutex.Lock()
	tree.RestartNodesWithFeature(file)
	fileMutex.Unlock()
}

func startWatchingFile(file string) {
	_, err := watcherIn.Write([]byte(file + "\n"))
	if err != nil {
		slog.Error(err)
	}
}
