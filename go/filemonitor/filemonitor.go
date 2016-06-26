package filemonitor

import (
	"sync"
	"time"
)

const DefaultFileChangeDelay = 300 * time.Millisecond

type FileMonitor interface {
	Listen() <-chan []string
	Add(string) error
	Close() error
}

type fileMonitor struct {
	listeners     []chan []string
	listenerMutex sync.Mutex
}

func (f *fileMonitor) Listen() <-chan []string {
	f.listenerMutex.Lock()
	defer f.listenerMutex.Unlock()

	c := make(chan []string)
	f.listeners = append(f.listeners, c)

	return c
}

type gatheringMonitor struct {
	fileMonitor
	changes         chan string
	fileChangeDelay time.Duration
}

// Create the changes channel and serve debounced changes to listeners.
// The changes channel must be created before this is started.
// Closing the changes channel causes this to close all listener
// channels and return.
func (f *gatheringMonitor) serveListeners() {
	never := make(<-chan time.Time)
	deadline := never

	collected := make(map[string]bool, 1)
	for {
		select {
		case change := <-f.changes:
			// Channel closed
			if change == "" {
				f.listenerMutex.Lock()
				defer f.listenerMutex.Unlock()

				for _, listener := range f.listeners {
					close(listener)
				}
				return
			}

			collected[change] = true
			if deadline == never {
				deadline = time.After(f.fileChangeDelay)
			}
		case <-deadline:
			list := make([]string, 0, len(collected))
			for f := range collected {
				list = append(list, f)
			}

			for _, l := range f.listeners {
				l <- list
			}

			deadline = never
			collected = make(map[string]bool, 1)
		}
	}
}
