// +build darwin

package filemonitor

import (
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsevents"
)

const flagsWorthReloadingFor = fsevents.ItemRemoved | fsevents.ItemModified | fsevents.ItemRenamed

type fsEventsMonitor struct {
	fileMonitor
	stream *fsevents.EventStream
	add    chan string
	stop   chan struct{}
}

func NewFileMonitor(fileChangeDelay time.Duration) (FileMonitor, error) {
	f := fsEventsMonitor{
		stream: &fsevents.EventStream{
			Paths:   []string{},
			Latency: fileChangeDelay,
			Flags:   fsevents.FileEvents,
			EventID: uint64(0xFFFFFFFFFFFFFFFF),
		},
		// Restarting FSEvents can take ~100ms so buffer adds
		// in the channel so they can be grouped together.
		add:  make(chan string, 5000),
		stop: make(chan struct{}),
	}

	go f.handleAdd()

	return &f, nil
}

func (f *fsEventsMonitor) Add(file string) error {
	f.add <- file
	return nil
}

func (f *fsEventsMonitor) Close() error {
	select {
	case <-f.stop:
		return nil // Already stopped
	default:
		close(f.stop)
		close(f.add)
	}

	return nil
}

func (f *fsEventsMonitor) watch() {
	for {
		select {
		case events := <-f.stream.Events:
			paths := make([]string, 0, len(events))
			for _, event := range events {
				if (event.Flags & (fsevents.ItemIsFile | flagsWorthReloadingFor)) == 0 {
					continue
				}

				paths = append(paths, event.Path)
			}

			if len(paths) == 0 {
				continue
			}

			for _, l := range f.listeners {
				l <- paths
			}
		case <-f.stop:
			return
		}
	}
}

func (f *fsEventsMonitor) handleAdd() {
	watched := make(map[string]bool)
	started := false
	// We don't want to add individual files to watch here but figure out the
	// directory to watch and watch it instead.

	for file := range f.add {
		path, err := pathToMonitor(file)
		if err != nil {
			// can't access the file for some reason, best to ignore, probably should
			// log something here
			continue
		}

		if watched[path] {
			continue
		}

		allFiles := []string{path}

		// Read all messages waiting in the channel so we can batch restarts
		done := false
		for !done {
			select {
			case file := <-f.add:
				path, err := pathToMonitor(file)
				if err != nil {
					continue
				}

				if watched[path] {
					continue
				}

				allFiles = append(allFiles, path)
			default:
				done = true
			}
		}

		for _, file := range allFiles {
			watched[file] = true
			if contains(f.stream.Paths, file) {
				continue
			}

			f.stream.Paths = append(f.stream.Paths, file)
		}

		if started {
			f.stream.Restart()
		} else {
			f.stream.Start()
			go f.watch()
			started = true
		}
	}

	if started {
		f.stream.Stop()
	}
}

func pathToMonitor(rawPath string) (path string, err error) {
	stat, err := os.Stat(rawPath)
	if err != nil {
		return
	}

	if stat.IsDir() {
		path = rawPath
	} else {
		path = filepath.Dir(rawPath)
	}

	return
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}

	return false
}
