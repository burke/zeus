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

func Run() {
	os.Exit(doRun())
}

func doRun() int {
	slog.Colorized("{green}Starting {yellow}Z{red}e{blue}u{magenta}s{green} server")

	exitNow = make(chan int)

	var tree *ProcessTree = BuildProcessTree()

	done := make(chan bool)

	quit1 := StartSlaveMonitor(tree, done)
	defer func() { <-done }()

	quit2 := StartClientHandler(tree, done)
	defer func() { <-done }()

	quit3 := StartFileMonitor(tree, done)
	defer func() { <-done }()

	quit4 := StartStatusChart(tree, done)
	defer func() { <-done }()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		<-c
		slog.Suppress()
		go func() { quit1 <- true }()
		go func() { quit2 <- true }()
		go func() { quit3 <- true }()
		go func() { quit4 <- true }()
	}()

	go func() {
		exitStatus = <-exitNow
		slog.Suppress()
		go func() { quit1 <- true }()
		go func() { quit2 <- true }()
		go func() { quit3 <- true }()
		go func() { quit4 <- true }()
	}()

	return exitStatus
}
