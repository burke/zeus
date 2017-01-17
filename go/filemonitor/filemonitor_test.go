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

func writeTestFiles(dir string, count int) ([]string, error) {
	files := make([]string, count)

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

func TestFileMonitorManyFiles(t *testing.T) {
	count := 5000
	dir, err := ioutil.TempDir("", "zeus_test_many_files")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	files, err := writeTestFiles(dir, count)
	if err != nil {
		t.Fatal(err)
	}

	fm, err := filemonitor.NewFileMonitor(filemonitor.DefaultFileChangeDelay)
	if err != nil {
		t.Fatal(err)
	}
	defer fm.Close()

	watched := files[0:count]
	for i, file := range watched {
		err = fm.Add(file)
		if err != nil {
			t.Fatalf("%d, %v", i, err)
		}
	}

	changes := fm.Listen()

	// Without a short sleep we get notified for the original
	// file creation using FSEvents
	time.Sleep(20 * time.Millisecond)

	for i, file := range files {
		if err := ioutil.WriteFile(file, []byte("baz"), 0); err != nil {
			t.Fatalf("%d: %v", i, err)
		}
	}

	if err := expectChanges(changes, watched); err != nil {
		t.Fatal(err)
	}
}

func TestFileMonitor(t *testing.T) {
	dir, err := ioutil.TempDir("", "zeus_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	files, err := writeTestFiles(dir, 3)
	if err != nil {
		t.Fatal(err)
	}

	fm, err := filemonitor.NewFileMonitor(filemonitor.DefaultFileChangeDelay)
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
}

func expectChanges(changeCh <-chan []string, expect []string) error {
	// Copy the input before sorting
	expectSorted := make([]string, len(expect))
	copy(expectSorted, expect)
	sort.StringSlice(expectSorted).Sort()
	expect = expectSorted

	select {
	case changes := <-changeCh:
		sort.StringSlice(changes).Sort()

		if !reflect.DeepEqual(changes, expect) {
			return fmt.Errorf("expected changes in %v, got %v", expect, changes)
		}
	case <-time.After(time.Second):
		return errors.New("Timeout waiting for change notification")
	}

	return nil
}
