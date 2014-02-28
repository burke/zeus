package restarter

import (
	"github.com/burke/zeus/go/processtree"
	slog "github.com/burke/zeus/go/shinylog"
	"time"
)

var FileChangeWindow = 300 * time.Millisecond

func Start(tree *processtree.ProcessTree, filesChanged chan string, done chan bool) (quit chan bool) {
	quit = make(chan bool)
	rst := &restarter{tree, filesChanged, done, quit}
	go rst.start()
	return quit
}

type restarter struct {
	tree         *processtree.ProcessTree
	filesChanged chan string
	done, quit   chan bool
}

// Collect file changes that happen FileChangeWindow ms from each
// other, and restart all nodes in the process tree that match
// features of the changed files.
func (r *restarter) start() {
	for {
		select {
		case <-r.quit:
			r.done <- true
			return
		case file := <-r.filesChanged:
			changed := make(map[string]bool)
			changed[file] = true

			slog.Trace("Restarter got the first file of potentially many")

			if r.gatherFiles(changed, time.After(FileChangeWindow)) {
				return
			}

			slog.Trace("Restarter has gathered %d changed files", len(changed))
			go r.tree.RestartNodesWithFeatures(changed)
		}
	}
}

func (r *restarter) gatherFiles(changed map[string]bool, deadline <-chan time.Time) (quitNow bool) {
	for {
		select {
		case <-r.quit:
			r.done <- true
			return true
		case file := <-r.filesChanged:
			changed[file] = true
		case <-deadline:
			return false
		}
	}
}
