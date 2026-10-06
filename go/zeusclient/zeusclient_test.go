package zeusclient

import (
	"io/ioutil"
	"os"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/burke/zeus/go/unixsocket"
)

// Regression test: the master may be slow to pick up the args fd sent via
// SCM_RIGHTS. On macOS, if the client has already closed its copy of that
// socket while the fd is still in transit, the arguments buffered in it are
// discarded and the master reads nothing. The client must therefore hold on
// to its copy until the master has acknowledged receipt.
func TestSendCommandLineArgumentsSlowReceiver(t *testing.T) {
	for i := 0; i < 20; i++ {
		clientSide, masterSide, err := unixsocket.Socketpair(syscall.SOCK_STREAM)
		if err != nil {
			t.Fatal(err)
		}
		clientUsock, err := unixsocket.NewFromFile(clientSide)
		if err != nil {
			t.Fatal(err)
		}
		masterUsock, err := unixsocket.NewFromFile(masterSide)
		if err != nil {
			t.Fatal(err)
		}

		args := []string{"zeus", "rspec", "spec/foo_spec.rb:12"}
		remoteArgs, err := sendCommandLineArguments(clientUsock, args)
		if err != nil {
			t.Fatal(err)
		}

		// Master is slow to wake up and read the fd.
		time.Sleep(50 * time.Millisecond)

		argFD, err := masterUsock.ReadFD()
		if err != nil {
			t.Fatal(err)
		}
		argFile := os.NewFile(uintptr(argFD), "args")
		data, err := ioutil.ReadAll(argFile)
		if err != nil {
			t.Fatal(err)
		}
		argFile.Close()

		// Master has received everything; the client may now drop its copy.
		remoteArgs.Close()

		got := strings.Split(strings.TrimRight(string(data), "\000"), "\000")
		want := args[1:]
		if strings.Join(got, "\000") != strings.Join(want, "\000") {
			t.Fatalf("iteration %d: got args %q, want %q", i, got, want)
		}

		clientUsock.Close()
		masterUsock.Close()
	}
}
