package unixsocket

import (
	"io/ioutil"
	"os"
	"strings"
	"syscall"
	"testing"
)

func TestLongMessage(t *testing.T) {
	message := strings.Repeat("abcdefghijklmonpqrstuvwxyz", 1000)

	a, b := makeUsockPair(t)

	go sendMessage(t, a, message)
	expectMessage(t, b, message)
}

func TestMessagesAndFDs(t *testing.T) {
	var msg string
	a, b := makeUsockPair(t)

	tempFile := makeTempFile(t)
	defer os.Remove(tempFile.Name())

	messages := []string{"zomg", "wtf", "lol"}

	sendFD(t, a, tempFile.Fd())
	for _, msg = range messages {
		sendMessage(t, a, msg)
	}

	for _, msg = range messages {
		expectMessage(t, b, msg)
	}
	expectFD(t, b, tempFile.Fd())
}

func makeUsockPair(t *testing.T) (sockA, sockB *Usock) {
	a, b, err := Socketpair(syscall.SOCK_STREAM)
	if err != nil {
		t.Fatal(err)
	}

	sockA, err = NewFromFile(a)
	if err != nil {
		t.Fatal(err)
	}

	sockB, err = NewFromFile(b)
	if err != nil {
		t.Fatal(err)
	}

	return
}

func makeTempFile(t *testing.T) (tempFile *os.File) {
	tempFile, err := ioutil.TempFile("/tmp", "unixsocket_test")
	if err != nil {
		t.Fatal(err)
	}
	return
}

func expectMessage(t *testing.T, b *Usock, msg string) {
	readMsg, err := b.ReadMessage()
	if err != nil {
		t.Error(err)
	}
	if readMsg != msg {
		t.Errorf("Expected \"%s\", but read \"%s\"\n", msg, readMsg)
	}
}

func sendMessage(t *testing.T, a *Usock, msg string) {
	n, err := a.WriteMessage(msg)
	if err != nil {
		t.Error(err)
	}
	if n != len(msg) {
		t.Errorf("Expected %d bytes written, but was %d\n", len(msg), n)
	}
}

func sendFD(t *testing.T, a *Usock, fd uintptr) {
	if err := a.WriteFD(int(fd)); err != nil {
		t.Error(err)
	}
}

func expectFD(t *testing.T, b *Usock, compareFd uintptr) {
	fd, err := b.ReadFD()
	if err != nil {
		t.Error(err)
	}
	if fd <= int(compareFd) {
		t.Errorf("Expected new FD, but got %d\n", fd)
	}
}
