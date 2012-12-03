package zeusmaster

import (
	"os"
	"os/signal"

	slog "github.com/burke/zeus/go/shinylog"
)

var exitNow chan int

func ExitNow(code int) {
	exitNow <- code
}

func Run() {
	os.Exit(doRun())
}

func doRun() int {
	slog.Colorized("{green}Starting {yellow}Z{red}e{blue}u{magenta}s{green} server")

	var tree *ProcessTree = BuildProcessTree()

	done := make(chan bool)
	// Start processes and register them for exit when the function returns.
	defer exit(StartSlaveMonitor(tree, done), done)
	defer exit(StartClientHandler(tree, done), done)
	defer exit(StartFileMonitor(tree, done), done)
	defer exit(StartStatusChart(tree, done), done)
	defer slog.Suppress()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	select {
	case <-c:
		return 0
	case exitStatus := <-exitNow:
		return exitStatus
	}
}

func exit(quit, done chan bool) {
	// Signal the process to quit.
	quit <- true
	// Wait until the process signals it's done.
	<-done
}

func init() {
	exitNow = make(chan int)
}
