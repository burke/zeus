package process_test

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/burke/zeus/go/processtree/process"
)

func start(plan string) (process.NodeProcess, error) {
	gempath := os.Getenv("ZEUS_TEST_GEMPATH")
	if gempath == "" {
		var err error
		gempath, err = filepath.Abs("rubygem")
		if err != nil {
			return nil, fmt.Errorf("error constructing default gem path: %v", err)
		}
	}

	libpath := filepath.Join(gempath, "lib")
	if _, err := os.Stat(libpath); err != nil {
		return nil, fmt.Errorf("invalid gem lib path %s: %v", libpath, err)
	}

	args := []string{
		"ruby",
		fmt.Sprintf("-e$:.unshift %q", libpath),
		"-erequire 'zeus'",
		fmt.Sprintf(`-eclass MyPlan < Zeus::Plan
%s
end
Zeus.plan = MyPlan.new
`, plan),
		"-eZeus.go",
	}

	proc, err := process.StartProcess(args)
	if err != nil {
		return nil, err
	}

	return proc, nil
}

func startSuccessfully(plan string) (process.NodeProcess, error) {
	proc, err := start(plan)
	if err != nil {
		return nil, err
	}

	select {
	case <-proc.Ready():
		return proc, nil
	case err := <-proc.Errors():
		if err := proc.Stop(); err != nil {
			return nil, err
		}
		return nil, err
	}
}

func bootChild(parent process.NodeProcess, name string) (process.NodeProcess, error) {
	req := process.NewNodeRequest("foo")
	parent.Boot(req.BootRequest)
	child, err := req.Await()
	if err != nil {
		return nil, err
	}

	return child, nil
}

func TestStartZeusProcess(t *testing.T) {
	proc, err := startSuccessfully("def boot; end")
	if err != nil {
		t.Fatal(err)
	}

	files := proc.Files()

	defer func() {
		if err := proc.Stop(); err != nil {
			t.Fatal(err)
		}

		select {
		case f, ok := <-files:
			if ok {
				t.Errorf("files channel should be closed after stopping but it returned %q", f)
			}
		default:
			t.Errorf("files channel should be closed after stopping but it blocked")
		}
	}()

	select {
	case f := <-files:
		t.Errorf("did not expect to read any files, got %q", f)
	default:
	}

	if have, want := proc.Name(), "boot"; have != want {
		t.Errorf("expected identifier %q, got %q", want, have)
	}
}

func TestStartZeusProcessError(t *testing.T) {
	proc, err := start("")
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := proc.Stop(); err != nil {
			t.Fatal(err)
		}
	}()

	select {
	case <-proc.Ready():
		t.Fatal("process booted successfully, expected an error")
	case err := <-proc.Errors():
		if have, want := err.Error(), "undefined method `boot'"; !strings.Contains(have, want) {
			t.Errorf("expected error to contain %q, got %v", want, have)
		}
	}

}

func TestBootNode(t *testing.T) {
	proc, err := startSuccessfully("def boot; end; def foo; end")
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := proc.Stop(); err != nil {
			t.Fatal(err)
		}
	}()

	child, err := bootChild(proc, "foo")
	if err != nil {
		t.Fatal(err)
	}

	if have, want := child.ParentPid(), proc.Pid(); have != want {
		t.Errorf("child thinks its parent is %d but should be %d", have, want)
	}

	if have, want := child.Name(), "foo"; have != want {
		t.Errorf("expected identifier %q, got %q", want, have)
	}

	defer func() {
		if err := child.Stop(); err != nil {
			t.Fatal(err)
		}
	}()

	select {
	case <-child.Ready():
	case err := <-child.Errors():
		t.Fatal(err)
	}
}

func TestBootNodeError(t *testing.T) {
	proc, err := startSuccessfully("def boot; end;")
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := proc.Stop(); err != nil {
			t.Fatal(err)
		}
	}()

	child, err := bootChild(proc, "foo")
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := child.Stop(); err != nil {
			t.Fatal(err)
		}
	}()

	select {
	case <-child.Ready():
		t.Fatal("process booted successfully, expected an error")
	case err := <-child.Errors():
		if have, want := err.Error(), "undefined method `foo'"; !strings.Contains(have, want) {
			t.Errorf("expected error to contain %q, got %v", want, have)
		}
	}
}

func TestBootNodeAfterStop(t *testing.T) {
	proc, err := startSuccessfully("def boot; end; def foo; end")
	if err != nil {
		t.Fatal(err)
	}

	if err := proc.Stop(); err != nil {
		t.Fatal(err)
	}

	child, err := bootChild(proc, "foo")
	if err != process.ErrProcessStopping {
		t.Errorf("expected ErrProcessStopping booting after stopping, got %v", err)
	}
	if err == nil {
		if err := child.Stop(); err != nil {
			t.Fatal(err)
		}
	}
}

func TestFileTracking(t *testing.T) {
	trackFile, err := ioutil.TempFile("", "zeus-test-track")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(trackFile.Name())

	plan := fmt.Sprintf("def boot; Zeus::LoadTracking.add_feature(%q); end", trackFile.Name())
	proc, err := startSuccessfully(plan)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := proc.Stop(); err != nil {
			t.Fatal(err)
		}
	}()

	files := proc.Files()

	select {
	case f := <-files:
		if f != trackFile.Name() {
			t.Errorf("expected file %q, got %q", trackFile, f)
		}
	case <-time.After(time.Second):
		t.Errorf("no files reported before timeout")
	}
}
