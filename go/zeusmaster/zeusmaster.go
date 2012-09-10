package zeusmaster

import (
	"os"
	"os/signal"

	slog "github.com/burke/zeus/go/shinylog"
)

var exitNow chan int
var exitStatus int = 0

func ExitNow(code int) {
	exitNow <- code
}

func Run(color bool) {
	exitNow = make(chan int)

	if !color {
		slog.DisableColor()
	}
	startingZeus()

	var tree *ProcessTree = BuildProcessTree()

	quitters := []chan bool{make(chan bool), make(chan bool), make(chan bool)}

	go StartSlaveMonitor(tree, quitters[0])
	go StartClientHandler(tree, quitters[1])
	go StartFileMonitor(tree, quitters[2])

	quit := make(chan bool, 1)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		terminateComponents(quitters, quit)
	}()

	go func() {
		exitStatus = <-exitNow
		terminateComponents(quitters, quit)
	}()

	for _, _ = range quitters {
		<-quit
	}

	os.Exit(exitStatus)
}

func terminateComponents(quitters []chan bool, quit chan bool) {
	slog.Suppress()
	for _, quitter := range quitters {
		go func() {
			quitter <- true
			<-quitter
			quit <- true
		}()
	}
}

func startingZeus() {
	slog.Colorized("{green}Starting {yellow}Z{red}e{blue}u{magenta}s{green} server")
}
