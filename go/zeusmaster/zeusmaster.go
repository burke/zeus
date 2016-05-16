package zeusmaster

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/burke/zeus/go/clienthandler"
	"github.com/burke/zeus/go/config"
	"github.com/burke/zeus/go/filemonitor"
	"github.com/burke/zeus/go/processtree"
	"github.com/burke/zeus/go/restarter"
	slog "github.com/burke/zeus/go/shinylog"
	"github.com/burke/zeus/go/statuschart"
	"github.com/burke/zeus/go/zerror"
	"github.com/burke/zeus/go/zeusversion"
)

// man signal | grep 'terminate process' | awk '{print $2}' | xargs -I '{}' echo -n "syscall.{}, "
// Leaving out SIGPIPE as that is a signal the master receives if a client process is killed.
var terminatingSignals = []os.Signal{syscall.SIGHUP, syscall.SIGINT, syscall.SIGKILL, syscall.SIGALRM, syscall.SIGTERM, syscall.SIGXCPU, syscall.SIGXFSZ, syscall.SIGVTALRM, syscall.SIGPROF, syscall.SIGUSR1, syscall.SIGUSR2}

func Run() int {
	slog.Colorized("{green}Starting {yellow}Z{red}e{blue}u{magenta}s{green} server v" + zeusversion.VERSION)

	zerror.Init()
	var tree *processtree.ProcessTree = config.BuildProcessTree()

	done := make(chan bool)

	// Start processes and register them for exit when the function returns.
	filesChanged, filemonitorDone := filemonitor.Start(done)

	defer exit(processtree.StartSlaveMonitor(tree, done), done)
	defer exit(clienthandler.Start(tree, done), done)
	defer exit(filemonitorDone, done)
	defer exit(restarter.Start(tree, filesChanged, done), done)
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
		}
	}
}

func exit(quit, done chan bool) {
	// Signal the process to quit.
	close(quit)
	// Wait until the process signals it's done.
	<-done
}
