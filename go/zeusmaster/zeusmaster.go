package zeusmaster

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/burke/zeus/go/clienthandler"
	"github.com/burke/zeus/go/config"
	"github.com/burke/zeus/go/filemonitor"
	"github.com/burke/zeus/go/processtree"
	slog "github.com/burke/zeus/go/shinylog"
	"github.com/burke/zeus/go/statuschart"
	"github.com/burke/zeus/go/zerror"
)

// man signal | grep 'terminate process' | awk '{print $2}' | xargs -I '{}' echo -n "syscall.{}, "
var terminatingSignals = []os.Signal{syscall.SIGHUP, syscall.SIGINT, syscall.SIGKILL, syscall.SIGPIPE, syscall.SIGALRM, syscall.SIGTERM, syscall.SIGXCPU, syscall.SIGXFSZ, syscall.SIGVTALRM, syscall.SIGPROF, syscall.SIGUSR1, syscall.SIGUSR2}

func Run() {
	zerror.Init()
	os.Exit(doRun())
}

func doRun() int {
	slog.Colorized("{green}Starting {yellow}Z{red}e{blue}u{magenta}s{green} server")

	var tree *processtree.ProcessTree = config.BuildProcessTree()

	done := make(chan bool)

	// Start processes and register them for exit when the function returns.
	filesChanged, filemonitorDone := filemonitor.Start(done)

	defer exit(processtree.StartSlaveMonitor(tree, done), done)
	defer exit(clienthandler.Start(tree, done), done)
	defer exit(filemonitorDone, done)
	defer slog.Suppress()
	defer zerror.PrintFinalOutput()
	defer exit(statuschart.Start(tree, done), done)

	c := make(chan os.Signal, 1)
	signal.Notify(c, terminatingSignals...)

	for {
		select {
		case sig := <-c:
			if sig == syscall.SIGINT {
				return 0
			} else {
				return 1
			}
		case changed := <-filesChanged:
			go tree.RestartNodesWithFeature(changed)
		}
	}
	return -1 // satisfy the compiler
}

func exit(quit, done chan bool) {
	// Signal the process to quit.
	quit <- true
	// Wait until the process signals it's done.
	<-done
}
