package zeusmaster_test

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"

	"github.com/burke/zeus/go/filemonitor"
	slog "github.com/burke/zeus/go/shinylog"
	"github.com/burke/zeus/go/unixsocket"
	"github.com/burke/zeus/go/zeusclient"
	"github.com/burke/zeus/go/zeusmaster"
)

var testFiles = map[string]string{
	"zeus.json": `
{
  "command": "ruby -r./custom_plan -eZeus.go",
  "plan": {
    "boot": {
      "data": {
        "data_srv": {}
      },
      "code": {
        "code_srv": {}
      },
      "cmd": []
    }
  }
}
`,
	"custom_plan.rb": `
$LOAD_PATH.unshift(File.readlink('./lib'))
require 'zeus'

class CustomPlan < Zeus::Plan
  def self.command(name, &block)
    define_method(name) do
      begin
        self.instance_eval(&block)
      rescue => e
        STDERR.puts "#{name} terminated with exception: #{e.message}"
        STDERR.puts e.backtrace.map {|line| " #{line}"}
        raise
      end
    end
  end

  command :boot do
    require_relative 'srv'
  end

  command :data do
    redirect_log('data')
    require_relative 'data'
  end

  command :code do
    redirect_log('code')
    require_relative 'code'
  end

  command :cmd do
    puts "bijagua"
    STDERR.puts "bazinga"
  end

  command :data_srv do
    redirect_log('data_srv')
    serve('data.sock')
  end

  command :code_srv do
    redirect_log('code_srv')
    serve('code.sock')
  end

  def redirect_log(cmd)
    log_file = File.open("zeus_test_#{cmd}.log", 'a')
    log_file.sync = true
    STDOUT.reopen(log_file)
    STDERR.reopen(log_file)
    STDOUT.sync = STDERR.sync = true
  end
end

Zeus.plan = CustomPlan.new
`,
	"data.rb": `
require 'yaml'
$response = YAML::load_file('data.yaml')['response']
`,
	"data.yaml": `
response: YAML the Camel is a Mammal with Enamel
`,
	"other-data.yaml": `
response: Hi
`,
	"code.rb": `
$response = "Hello, world!"
`,
	"other-code.rb": `
$response = "there!"
`,
	"srv.rb": `
$response = "pong"

def serve(sock_path)
  sock = Socket.new(Socket::AF_UNIX, Socket::SOCK_DGRAM, 0)
  sock.connect(Socket.pack_sockaddr_un(sock_path))

  b = sock.send($response, 0)
  puts "Wrote #{b} bytes to #{sock_path}"
end
`,
}

func writeTestFiles(dir string) error {
	for name, contents := range testFiles {
		if err := ioutil.WriteFile(path.Join(dir, name), []byte(contents), 0644); err != nil {
			return fmt.Errorf("error writing %s: %v", name, err)
		}
	}

	gempath := os.Getenv("ZEUS_TEST_GEMPATH")
	if gempath == "" {
		var err error
		gempath, err = filepath.Abs("rubygem")
		if err != nil {
			return fmt.Errorf("error finding gempath: %v", err)
		}
	}

	if err := os.Symlink(filepath.Join(gempath, "lib"), filepath.Join(dir, "lib")); err != nil {
		return fmt.Errorf("error linking zeus gem: %v", err)
	}

	return nil
}

func enableTracing() {
	slog.SetTraceLogger(slog.NewTraceLogger(os.Stderr))
}

func TestZeusBoots(t *testing.T) {
	dir, err := ioutil.TempDir("", "zeus_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	if err := writeTestFiles(dir); err != nil {
		t.Fatal(err)
	}

	unixsocket.SetZeusSockName(filepath.Join(dir, ".zeus.sock"))

	connections := map[string]*net.UnixConn{
		"cmd":  nil,
		"data": nil,
		"code": nil,
	}

	for name := range connections {
		sockName := filepath.Join(dir, fmt.Sprintf("%s.sock", name))

		c, err := net.ListenUnixgram("unixgram", &net.UnixAddr{
			Name: sockName, Net: "unixgram",
		})
		if err != nil {
			t.Fatalf("Error opening %q socket: %v", sockName, err)
		}
		defer c.Close()

		connections[name] = c
	}

	me, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatal(err)
	}
	defer me.Signal(os.Interrupt)

	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	// TODO: Find a way to redirect stdout so we can look for crashed
	// processes.
	enableTracing()
	zexit := make(chan int)
	go func() {
		zexit <- zeusmaster.Run(filepath.Join(dir, "zeus.json"), filemonitor.DefaultFileChangeDelay)
	}()

	expects := map[string]string{
		// TODO: Use the zeusclient to spawn a command to test
		// that path.
		// "cmd":  "pong",
		"data": "YAML the Camel is a Mammal with Enamel",
		"code": "Hello, world!",
	}

	for name, want := range expects {
		if err := readAndCompare(connections[name], want); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
	}

	time.Sleep(400 * time.Millisecond)

	for _, f := range []string{"code.rb", "data.yaml"} {
		from := filepath.Join(dir, fmt.Sprintf("other-%s", f))
		to := filepath.Join(dir, f)
		if err := os.Rename(from, to); err != nil {
			t.Fatalf("Error renaming %s: %v", f, err)
		}
	}

	expects = map[string]string{
		"data": "Hi",
		"code": "there!",
	}

	for name, want := range expects {
		if err := readAndCompare(connections[name], want); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
	}

	readCloser := make(chan struct{})
	defer func() { close(readCloser) }()

	cmdReader, cmdWriter, err := os.Pipe()

	if err != nil {
		t.Fatal(err)
	}

	cmdErrReader, cmdErrWriter, err := os.Pipe()

	if err != nil {
		t.Fatal(err)
	}

	cexit := make(chan int, 1)
	go func() {
		cexit <- zeusclient.Run([]string{"cmd"}, hangingReader{readCloser}, cmdWriter, cmdErrWriter)
		time.Sleep(100 * time.Millisecond)
		cmdWriter.Close()
		cmdErrWriter.Close()
	}()

	have, err := ioutil.ReadAll(cmdReader)
	if err != nil {
		t.Fatal(err)
	}
	if want := "bijagua\n"; string(have) != want {
		t.Errorf("expected %q, got %q", want, have)
	}
	if code := <-cexit; code != 0 {
		t.Errorf("cmd exited with %d", code)
	}

	have, err = ioutil.ReadAll(cmdErrReader)
	if err != nil {
		t.Fatal(err)
	}
	if want := "bazinga\n"; string(have) != want {
		t.Errorf("expected stderr %q, got %q", want, have)
	}

	// The zeusmaster catches the interrupt and exits gracefully
	me.Signal(os.Interrupt)
	if code := <-zexit; code != 0 {
		t.Fatalf("Zeus exited with %d", code)
	}
}

func readAndCompare(conn *net.UnixConn, want string) error {
	buf := make([]byte, 128)

	// File system events can take a long time to propagate
	conn.SetDeadline(time.Now().Add(2 * time.Second))

	if _, _, err := conn.ReadFrom(buf); err != nil {
		return err
	}
	if have := string(bytes.Trim(buf, "\x00")); have != want {
		return fmt.Errorf("expected %q, got %q", want, have)
	}

	return nil
}

type hangingReader struct {
	close chan struct{}
}

func (r hangingReader) Read([]byte) (int, error) {
	<-r.close
	return 0, io.EOF
}
