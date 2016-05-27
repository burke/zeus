// +build !darwin

package filemonitor

import "github.com/fsnotify/fsnotify"

type fsnotifyMonitor struct {
	gatheringMonitor
	watcher *fsnotify.Watcher
}

const flagsWorthReloadingFor = fsnotify.Write | fsnotify.Remove | fsnotify.Rename

func NewFileMonitor() (FileMonitor, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	f := fsnotifyMonitor{
		watcher: watcher,
	}
	f.changes = make(chan string)

	go f.serveListeners()
	go f.watch()

	return &f, nil
}

func (f *fsnotifyMonitor) Add(file string) error {
	if err := f.watcher.Add(file); err != nil {
		return err
	}

	return nil
}

func (f *fsnotifyMonitor) Close() error {
	return f.watcher.Close()
}

func (f *fsnotifyMonitor) watch() {
	// Detect zero value and return
	// otherwise debounce

	for event := range f.watcher.Events {
		if (event.Op & flagsWorthReloadingFor) == 0 {
			continue
		}

		f.changes <- event.Name
	}

	close(f.changes)
}
