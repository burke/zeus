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

	quit1 := make(chan bool)
	quit2 := make(chan bool)
	quit3 := make(chan bool)
	quit4 := make(chan bool)

	go StartSlaveMonitor(tree, quit1)
	go StartClientHandler(tree, quit2)
	go StartFileMonitor(tree, quit3)
	go StartStatusChart(tree, quit4)

	quit := make(chan bool)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		terminateComponents(quit1, quit2, quit3, quit4, quit)
	}()

	go func() {
		exitStatus = <-exitNow
		terminateComponents(quit1, quit2, quit3, quit4, quit)
	}()

	<-quit
	<-quit
	<-quit
	<-quit

	os.Exit(exitStatus)
}

func terminateComponents(quit1, quit2, quit3, quit4, quit chan bool) {
	slog.Suppress()
	go func() {
		quit1 <- true
		<-quit1
		quit <- true
	}()
	go func() {
		quit2 <- true
		<-quit2
		quit <- true
	}()
	go func() {
		quit3 <- true
		<-quit3
		quit <- true
	}()
	go func() {
		quit4 <- true
		<-quit4
		quit <- true
	}()
}

func startingZeus() {
	slog.Colorized("{green}Starting {yellow}Z{red}e{blue}u{magenta}s{green} server")
}
