// +build darwin

package filemonitor

import (
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
			Paths: []string{},
			// We want debouncing rather than throttling so we need to handle
			// aggregating events ourselves
			Latency: 0,
			Flags:   fsevents.FileEvents,
			EventID: fsevents.EventIDSinceNow,
		},
		// Restarting FSEvents can take ~100ms so buffer adds
		// in the channel so they can be grouped together.
		add:  make(chan string, 5000),
		stop: make(chan struct{}),
	}
	f.fileChangeDelay = fileChangeDelay
	f.changes = make(chan string)

	go f.serveListeners()
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
			for _, event := range events {
				if (event.Flags & (fsevents.ItemIsFile | flagsWorthReloadingFor)) == 0 {
					continue
				}

				f.changes <- event.Path
			}
		case <-f.stop:
			close(f.changes)
			return
		}
	}
}

func (f *fsEventsMonitor) handleAdd() {
	watched := make(map[string]bool)
	started := false

	for file := range f.add {
		if watched[file] {
			continue
		}

		allFiles := []string{file}

		// Read all messages waiting in the channel so we can batch restarts
		done := false
		for !done {
			select {
			case file := <-f.add:
				if watched[file] {
					continue
				}

				allFiles = append(allFiles, file)
			default:
				done = true
			}
		}

		for _, file := range allFiles {
			watched[file] = true
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
