package restarter

import (
	"github.com/burke/zeus/go/processtree"
	slog "github.com/burke/zeus/go/shinylog"
	"time"
)

var FileChangeWindow = 300 * time.Millisecond

func Start(tree *processtree.ProcessTree, filesChanged chan string, done chan bool) (quit chan bool) {
	quit = make(chan bool)
	go start(tree, filesChanged, done, quit)
	return quit
}

// Collect file changes that happen FileChangeWindow ms from each
// other, and restart all nodes in the process tree that match
// features of the changed files.
func start(tree *processtree.ProcessTree, filesChanged chan string, done, quit chan bool) {
	for {
		select {
		case <-quit:
			done <- true
			return
		case file := <-filesChanged:
			changed := make(map[string]bool)
			changed[file] = true

			slog.Trace("Restarter got the first file of potentially many")
			deadline := time.After(FileChangeWindow)
			deadline_expired := false
			for !deadline_expired {
				select {
				case <-quit:
					done <- true
					return
				case file := <-filesChanged:
					changed[file] = true
				case <-deadline:
					deadline_expired = true
				}
			}
			slog.Trace("Restarter has gathered %d changed files", len(changed))
			go tree.RestartNodesWithFeatures(changed)
		}
	}
}
