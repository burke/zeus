package zeusmaster

import (
	"os"
	"os/signal"

	"github.com/burke/zeus/go/filemonitor"
	slog "github.com/burke/zeus/go/shinylog"
)

var exitNow chan int

var finalOutput []func()

func ExitNow(code int, finalOuputCallback func()) {
	finalOutput = append(finalOutput, finalOuputCallback)
	exitNow <- code
}

func Run() {
	finalOutput = make([]func(), 0)
	os.Exit(doRun())
}

func doRun() int {
	slog.Colorized("{green}Starting {yellow}Z{red}e{blue}u{magenta}s{green} server")

	var tree *ProcessTree = BuildProcessTree()

	done := make(chan bool)

	// Start processes and register them for exit when the function returns.
	filesChanged, filemonitorDone := filemonitor.Start(done)

	defer exit(StartSlaveMonitor(tree, done), done)
	defer exit(StartClientHandler(tree, done), done)
	defer exit(filemonitorDone, done)
	defer slog.Suppress()
	defer printFinalOutput()
	defer exit(StartStatusChart(tree, done), done)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	for {
		select {
		case <-c:
			return 0
		case exitStatus := <-exitNow:
			return exitStatus
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

func init() {
	exitNow = make(chan int)
}

func printFinalOutput() {
	for _, cb := range finalOutput {
		cb()
	}
}
