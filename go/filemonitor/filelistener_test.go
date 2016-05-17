package filemonitor_test

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"testing"
	"time"

	"github.com/burke/zeus/go/filemonitor"
	slog "github.com/burke/zeus/go/shinylog"
)

func TestFileListener(t *testing.T) {
	ln, err := net.ListenTCP("tcp", &net.TCPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 0,
	})
	if err != nil {
		t.Fatal(err)
	}

	slog.SetTraceLogger(slog.NewTraceLogger(os.Stderr))
	fl := filemonitor.NewFileListener(ln)
	defer fl.Close()

	// We should be able to add a file without connecting anything
	if err := fl.Add("foo"); err != nil {
		t.Fatal(err)
	}

	conn, err := net.DialTCP("tcp", nil, ln.Addr().(*net.TCPAddr))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	scanner := bufio.NewScanner(conn)

	// Can write files
	files := fl.Listen()
	if err := checkWrite(files, conn); err != nil {
		t.Fatal(err)
	}

	// Can read a file add operation
	want := "bar"
	if err := fl.Add(want); err != nil {
		t.Fatal(err)
	}
	conn.SetReadDeadline(time.Now().Add(time.Second))
	if err := checkScan(scanner, want); err != nil {
		t.Fatal(err)
	}

	// Can create a second connection
	conn2, err := net.DialTCP("tcp", nil, ln.Addr().(*net.TCPAddr))
	if err != nil {
		t.Fatal(err)
	}
	defer conn2.Close()
	scanner2 := bufio.NewScanner(conn2)

	// Can write a file to the second connection
	if err := checkWrite(files, conn2); err != nil {
		t.Fatal(err)
	}

	// Can read the same file Add from two connections
	want = "baz"
	if err := fl.Add(want); err != nil {
		t.Fatal(err)
	}

	conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	conn2.SetReadDeadline(time.Now().Add(100 * time.Millisecond))

	for i, s := range []*bufio.Scanner{scanner, scanner2} {
		if err := checkScan(s, want); err != nil {
			t.Errorf("%d: %v", i, err)
		}
	}

	// Can shutdown properly
	if err := fl.Close(); err != nil {
		t.Fatal(err)
	}

	for i, c := range []net.Conn{conn, conn2} {
		buf := make([]byte, 10)
		c.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
		if b, err := c.Read(buf); err == nil {
			t.Fatalf("%d: expected EOF reading closed connection but read %d bytes: %s", i, b, buf)
		} else if err != io.EOF {
			t.Fatalf("%d: expected EOF but got %v", i, err)
		}
	}
}

func checkScan(scanner *bufio.Scanner, want string) error {
	if scanner.Scan() {
		if have := scanner.Text(); have != want {
			return fmt.Errorf("expected %q, got %q", want, have)
		}
	} else {
		if err := scanner.Err(); err != nil {
			return err
		}
		return errors.New("Unexpected EOF")
	}

	return nil
}

func checkWrite(files <-chan []string, conn net.Conn) error {
	cases := make([]string, 2)
	for i := range cases {
		cases[i] = fmt.Sprintf("file%d", rand.Int())
	}

	for i, want := range cases {
		conn.SetWriteDeadline(time.Now().Add(100 * time.Millisecond))
		if _, err := conn.Write([]byte(want + "\n")); err != nil {
			return fmt.Errorf("Error writing %d: %v", i, err)
		}
	}

	if err := expectChanges(files, cases[:]); err != nil {
		return err
	}

	return nil
}
