package filemonitor

import (
	"bufio"
	"io"
	"net"
	"sync"
	"time"

	slog "github.com/burke/zeus/go/shinylog"
)

type fileListener struct {
	gatheringMonitor
	netListener net.Listener
	connections map[net.Conn]chan string
	stop        chan struct{}
	sync.Mutex
	wg sync.WaitGroup
}

func NewFileListener(ln net.Listener) FileMonitor {
	fl := fileListener{
		netListener: ln,
		connections: make(map[net.Conn]chan string),
		stop:        make(chan struct{}),
	}
	fl.changes = make(chan string)

	go fl.serveListeners()
	go fl.serve()

	return &fl
}

func (f *fileListener) Add(file string) error {
	f.Lock()
	defer f.Unlock()

	for _, ch := range f.connections {
		ch <- file
	}

	return nil
}

func (f *fileListener) Close() error {
	f.Lock()

	select {
	case <-f.stop:
		f.Unlock()
		return nil // Already stopped
	default:
		close(f.stop)
	}

	var firstErr error
	if firstErr = f.netListener.Close(); firstErr != nil {
		slog.Trace("Error closing file listener: %v", firstErr)
	}

	for conn := range f.connections {
		if err := conn.Close(); err != nil {
			if firstErr == nil {
				firstErr = err
			}
			slog.Trace("Error closing connection: %v", err)
		}
	}

	f.Unlock()
	f.wg.Wait()
	close(f.changes)

	return firstErr
}

func (f *fileListener) serve() {
	var tempDelay time.Duration // how long to sleep on accept failure

	for {
		conn, err := f.netListener.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				slog.Trace("filelistener: Accept error: %v; retrying in %v", err, tempDelay)
				time.Sleep(tempDelay)
				continue
			}

			select {
			case <-f.stop:
				return
			default:
				panic(err)
			}
		}

		ch := make(chan string)
		f.Lock()
		f.connections[conn] = ch
		f.wg.Add(1)
		f.Unlock()

		go f.handleConnection(conn, ch)
	}
}

func (f *fileListener) handleConnection(conn net.Conn, ch chan string) {
	// Handle writes
	stop := make(chan struct{})
	go func() {
		for {
			select {
			case s := <-ch:
				conn.SetWriteDeadline(time.Now().Add(1 * time.Second))
				if _, err := conn.Write([]byte(s + "\n")); err == io.EOF {
					return
				} else if err != nil {
					slog.Trace("Error writing to connection: %v", err)
				}
			case <-stop:
				return
			}
		}
	}()

	// Handle reads
	scanner := bufio.NewScanner(conn)
	for {
		if scanner.Scan() {
			f.changes <- scanner.Text()
		} else {
			if err := scanner.Err(); err != nil {
				select {
				case <-f.stop:
					break
				default:
					slog.Trace("Error reading from connection: %v", err)
				}
			}
			break
		}
	}

	f.Lock()
	defer f.Unlock()

	close(stop)
	delete(f.connections, conn)
	f.wg.Done()
}
