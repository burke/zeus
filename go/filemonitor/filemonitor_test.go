package filemonitor_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/burke/zeus/go/filemonitor"
)

// Setting a long delay here makes tests slow but improves reliability in Travis CI
const fileChangeDelay = 500 * time.Millisecond

func writeTestFiles(dir string) ([]string, error) {
	files := make([]string, 3)

	dir, err := filepath.EvalSymlinks(dir)
	if err != nil {
		return nil, err
	}

	for i := range files {
		fn := filepath.Join(dir, fmt.Sprintf("file%d", i))

		if err := ioutil.WriteFile(fn, []byte("foo"), 0644); err != nil {
			return nil, err
		}

		files[i] = fn
	}

	return files, nil
}

func TestFileMonitor(t *testing.T) {
	dir, err := ioutil.TempDir("", "zeus_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	files, err := writeTestFiles(dir)
	if err != nil {
		t.Fatal(err)
	}

	fm, err := filemonitor.NewFileMonitor(fileChangeDelay)
	if err != nil {
		t.Fatal(err)
	}
	defer fm.Close()

	changeCh := fm.Listen()

	// Without a short sleep we get notified for the original
	// file creation using FSEvents
	time.Sleep(20 * time.Millisecond)

	watched := files[0:2]
	for i, file := range watched {
		if err := fm.Add(file); err != nil {
			t.Fatalf("%d: %v", i, err)
		}
	}

	// Changing a file emits only that filename
	if err := ioutil.WriteFile(files[0], []byte("bar"), 0); err != nil {
		t.Fatal(err)
	}

	if err := expectChanges(changeCh, files[0:1]); err != nil {
		t.Fatal(err)
	}

	// Changing all files emits watched filenames together
	for i, file := range files {
		if err := ioutil.WriteFile(file, []byte("baz"), 0); err != nil {
			t.Fatalf("%d: %v", i, err)
		}
	}

	if err := expectChanges(changeCh, watched); err != nil {
		t.Fatal(err)
	}

	// Debouncing waits until no changes have occurred during the debounce
	// interval before reporting the change.
	for _, v := range [][]byte{{'1'}, {'2'}, {'3'}, {'4'}, {'5'}} {
		if err := ioutil.WriteFile(files[0], v, 0); err != nil {
			t.Fatal(err)
		}
		time.Sleep(fileChangeDelay / 3)
	}

	if err := expectChanges(changeCh, files[0:1]); err != nil {
		t.Fatal(err)
	}

	if changes := awaitChanges(changeCh); changes != nil {
		t.Fatalf("should not have any remaining changes, got %v", changes)
	}
}

func expectChanges(changeCh <-chan []string, expect []string) error {
	// Copy the input before sorting
	expectSorted := make([]string, len(expect))
	copy(expectSorted, expect)
	sort.StringSlice(expectSorted).Sort()
	expect = expectSorted

	changes := awaitChanges(changeCh)
	if changes == nil {
		return errors.New("Timeout waiting for change notification")
	}

	sort.StringSlice(changes).Sort()

	if !reflect.DeepEqual(changes, expect) {
		return fmt.Errorf("expected changes in %v, got %v", expect, changes)
	}

	return nil
}

func awaitChanges(changeCh <-chan []string) []string {
	select {
	case changes := <-changeCh:
		return changes
	case <-time.After(4 * fileChangeDelay):
		return nil
	}
}
