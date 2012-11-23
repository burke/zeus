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
	slog.Colorized("{green}Starting {yellow}Z{red}e{blue}u{magenta}s{green} server")

	exitNow = make(chan int)

	var tree *ProcessTree = BuildProcessTree()

	quit1 := StartSlaveMonitor(tree)
	quit2 := StartClientHandler(tree)
	quit3 := StartFileMonitor(tree)
	quit4 := StartStatusChart(tree)

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
