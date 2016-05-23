// +build darwin

// Package fsevents provides file system notifications on OS X.
package fsevents

/*
#cgo LDFLAGS: -framework CoreServices
#include <CoreServices/CoreServices.h>
#include <sys/stat.h>

static CFArrayRef ArrayCreateMutable(int len) {
	return CFArrayCreateMutable(NULL, len, &kCFTypeArrayCallBacks);
}

extern void fsevtCallback(FSEventStreamRef p0, uintptr_t info, size_t p1, char** p2, FSEventStreamEventFlags* p3, FSEventStreamEventId* p4);

static FSEventStreamRef EventStreamCreateRelativeToDevice(FSEventStreamContext * context, uintptr_t info, dev_t dev, CFArrayRef paths, FSEventStreamEventId since, CFTimeInterval latency, FSEventStreamCreateFlags flags) {
	context->info = (void*) info;
	return FSEventStreamCreateRelativeToDevice(NULL, (FSEventStreamCallback) fsevtCallback, context, dev, paths, since, latency, flags);
}

static FSEventStreamRef EventStreamCreate(FSEventStreamContext * context, uintptr_t info, CFArrayRef paths, FSEventStreamEventId since, CFTimeInterval latency, FSEventStreamCreateFlags flags) {
	context->info = (void*) info;
	return FSEventStreamCreate(NULL, (FSEventStreamCallback) fsevtCallback, context, paths, since, latency, flags);
}
*/
import "C"
import (
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

// EventIdSinceNow is a sentinel to begin watching events "since now".
const EventIDSinceNow = uint64(C.kFSEventStreamEventIdSinceNow + (1 << 64))

// CreateFlags for creating a New stream.
type CreateFlags uint32

// kFSEventStreamCreateFlag...
const (
	// use CoreFoundation types instead of raw C types (disabled)
	useCFTypes CreateFlags = 1 << iota

	// NoDefer sends events on the leading edge (for interactive applications).
	// By default events are delivered after latency seconds (for background tasks).
	NoDefer

	// WatchRoot for a change to occur to a directory along the path being watched.
	WatchRoot

	// IgnoreSelf doesn't send events triggered by the current process (OS X 10.6+).
	IgnoreSelf

	// FileEvents sends events about individual files, generating significantly
	// more events (OS X 10.7+) than directory level notifications.
	FileEvents
)

// EventFlags passed to the FSEventStreamCallback function.
type EventFlags uint32

// kFSEventStreamEventFlag...
const (
	// MustScanSubDirs indicates that events were coalesced hierarchically.
	MustScanSubDirs EventFlags = 1 << iota
	// UserDropped or KernelDropped is set alongside MustScanSubDirs
	// to help diagnose the problem.
	UserDropped
	KernelDropped

	// EventIDsWrapped indicates the 64-bit event ID counter wrapped around.
	EventIDsWrapped

	// HistoryDone is a sentinel event when retrieving events sinceWhen.
	HistoryDone

	// RootChanged indicates a change to a directory along the path being watched.
	RootChanged

	// Mount for a volume mounted underneath the path being monitored.
	Mount
	// Unmount event occurs after a volume is unmounted.
	Unmount

	// The following flags are only set when using FileEvents.

	ItemCreated
	ItemRemoved
	ItemInodeMetaMod
	ItemRenamed
	ItemModified
	ItemFinderInfoMod
	ItemChangeOwner
	ItemXattrMod
	ItemIsFile
	ItemIsDir
	ItemIsSymlink
)

// Event represents a single file system notification.
type Event struct {
	Path  string
	Flags EventFlags
	ID    uint64
}

//export fsevtCallback
func fsevtCallback(stream C.FSEventStreamRef, info uintptr, numEvents C.size_t, paths **C.char, flags *C.FSEventStreamEventFlags, ids *C.FSEventStreamEventId) {
	events := make([]Event, int(numEvents))

	es := registry.Get(info)
	if es == nil {
		return
	}

	for i := 0; i < int(numEvents); i++ {
		cpaths := uintptr(unsafe.Pointer(paths)) + (uintptr(i) * unsafe.Sizeof(*paths))
		cpath := *(**C.char)(unsafe.Pointer(cpaths))

		cflags := uintptr(unsafe.Pointer(flags)) + (uintptr(i) * unsafe.Sizeof(*flags))
		cflag := *(*C.FSEventStreamEventFlags)(unsafe.Pointer(cflags))

		cids := uintptr(unsafe.Pointer(ids)) + (uintptr(i) * unsafe.Sizeof(*ids))
		cid := *(*C.FSEventStreamEventId)(unsafe.Pointer(cids))

		events[i] = Event{Path: C.GoString(cpath), Flags: EventFlags(cflag), ID: uint64(cid)}
		// Record the latest EventID to support resuming the stream
		es.EventID = uint64(cid)
	}

	es.Events <- events
}

// LatestEventID returns the most recently generated event ID, system-wide.
func LatestEventID() uint64 {
	return uint64(C.FSEventsGetCurrentEventId())
}

// DeviceForPath returns the device ID for the specified volume.
func DeviceForPath(path string) (int32, error) {
	stat := syscall.Stat_t{}
	if err := syscall.Lstat(path, &stat); err != nil {
		return 0, err
	}
	return stat.Dev, nil
}

// EventIDForDeviceBeforeTime returns an event ID before a given time.
func EventIDForDeviceBeforeTime(dev int32, before time.Time) uint64 {
	tm := C.CFAbsoluteTime(before.Unix())
	return uint64(C.FSEventsGetLastEventIdForDeviceBeforeTime(C.dev_t(dev), tm))
}

/*

	Primary EventStream interface.
	You can provide your own event channel if you wish (or one will be created
	on Start).

	es := &EventStream{Paths: []string{"/tmp"}, Flags: 0}
	es.Start()
	es.Stop()

*/

// EventStream is the primary interface to FSEvents.
type EventStream struct {
	stream       C.FSEventStreamRef
	rlref        C.CFRunLoopRef
	hasFinalizer bool
	registryID   uintptr

	Events  chan []Event
	Paths   []string
	Flags   CreateFlags
	EventID uint64
	Resume  bool
	Latency time.Duration
	Device  int32
}

// eventStreamRegistry is a lookup table for EventStream references passed to
// cgo. In Go 1.6+ passing a Go pointer to a Go pointer to cgo is not allowed.
// To get around this issue, we pass only an integer.
type eventStreamRegistry struct {
	sync.Mutex
	m map[uintptr]*EventStream
	i uintptr
}

var registry = eventStreamRegistry{m: map[uintptr]*EventStream{}}

func (r *eventStreamRegistry) Add(e *EventStream) uintptr {
	r.Lock()
	defer r.Unlock()

	r.i++
	r.m[r.i] = e
	return r.i
}

func (r *eventStreamRegistry) Get(i uintptr) *EventStream {
	r.Lock()
	defer r.Unlock()

	return r.m[i]
}

func (r *eventStreamRegistry) Delete(i uintptr) {
	r.Lock()
	defer r.Unlock()

	delete(r.m, i)
}

func finalizer(es *EventStream) {
	// If an EventStream is freed without Stop being called it will
	// cause a panic. This avoids that, and closes the stream instead.
	es.Stop()
}

// Start listening to an event stream.
func (es *EventStream) Start() {
	cPaths := C.ArrayCreateMutable(C.int(len(es.Paths)))
	defer C.CFRelease(C.CFTypeRef(cPaths))

	for _, p := range es.Paths {
		p, _ = filepath.Abs(p)
		cpath := C.CString(p)
		defer C.free(unsafe.Pointer(cpath))

		str := C.CFStringCreateWithCString(nil, cpath, C.kCFStringEncodingUTF8)
		C.CFArrayAppendValue(cPaths, unsafe.Pointer(str))
	}

	since := C.FSEventStreamEventId(EventIDSinceNow)
	if es.Resume {
		since = C.FSEventStreamEventId(es.EventID)
	}

	if es.Events == nil {
		es.Events = make(chan []Event)
	}

	es.registryID = registry.Add(es)
	context := C.FSEventStreamContext{}
	info := C.uintptr_t(es.registryID)
	latency := C.CFTimeInterval(float64(es.Latency) / float64(time.Second))
	if es.Device != 0 {
		es.stream = C.EventStreamCreateRelativeToDevice(&context, info, C.dev_t(es.Device), cPaths, since, latency, C.FSEventStreamCreateFlags(es.Flags))
	} else {
		es.stream = C.EventStreamCreate(&context, info, cPaths, since, latency, C.FSEventStreamCreateFlags(es.Flags))
	}

	started := make(chan struct{})

	go func() {
		runtime.LockOSThread()
		es.rlref = C.CFRunLoopGetCurrent()
		C.FSEventStreamScheduleWithRunLoop(es.stream, es.rlref, C.kCFRunLoopDefaultMode)
		C.FSEventStreamStart(es.stream)
		close(started)
		C.CFRunLoopRun()
	}()

	if !es.hasFinalizer {
		runtime.SetFinalizer(es, finalizer)
		es.hasFinalizer = true
	}

	<-started
}

// Flush events that have occurred but haven't been delivered.
func (es *EventStream) Flush(sync bool) {
	if sync {
		C.FSEventStreamFlushSync(es.stream)
	} else {
		C.FSEventStreamFlushAsync(es.stream)
	}
}

// Stop listening to the event stream.
func (es *EventStream) Stop() {
	if es.stream != nil {
		C.FSEventStreamStop(es.stream)
		C.FSEventStreamInvalidate(es.stream)
		C.FSEventStreamRelease(es.stream)
		C.CFRunLoopStop(es.rlref)
		registry.Delete(es.registryID)
	}
	es.stream = nil
	es.registryID = 0
}

// Restart listening.
func (es *EventStream) Restart() {
	es.Stop()
	es.Resume = true
	es.Start()
}
