package zeusmaster

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/burke/zeus/go/clienthandler"
	"github.com/burke/zeus/go/config"
	"github.com/burke/zeus/go/filemonitor"
	"github.com/burke/zeus/go/processtree"
	slog "github.com/burke/zeus/go/shinylog"
	"github.com/burke/zeus/go/statuschart"
	"github.com/burke/zeus/go/zerror"
	"github.com/burke/zeus/go/zeusversion"
)

const listenerPortVar = "ZEUS_NETWORK_FILE_MONITOR_PORT"

// man signal | grep 'terminate process' | awk '{print $2}' | xargs -I '{}' echo -n "syscall.{}, "
// Leaving out SIGPIPE as that is a signal the master receives if a client process is killed.
var terminatingSignals = []os.Signal{syscall.SIGHUP, syscall.SIGINT, syscall.SIGKILL, syscall.SIGALRM, syscall.SIGTERM, syscall.SIGXCPU, syscall.SIGXFSZ, syscall.SIGVTALRM, syscall.SIGPROF, syscall.SIGUSR2}

const PidFile = ".zeus.pid"

func Run(configFile string, fileChangeDelay time.Duration, simpleStatus bool) int {
	slog.Colorized("{green}Starting {yellow}Z{red}e{blue}u{magenta}s{green} server v" + zeusversion.VERSION)

	zerror.Init()
	writePidFile()
	defer os.Remove(PidFile)

	c := make(chan os.Signal, 1)
	signal.Notify(c, terminatingSignals...)
	signal.Notify(c, syscall.SIGUSR1)

	for {
		code, restart := runSession(configFile, fileChangeDelay, simpleStatus, c)
		if !restart {
			slog.Suppress()
			zerror.PrintFinalOutput()
			return code
		}
		slog.Colorized("{green}Rebooting {yellow}Z{red}e{blue}u{magenta}s{green} server v" + zeusversion.VERSION)
		zerror.Init()
	}
}

func writePidFile() {
	os.WriteFile(PidFile, []byte(strconv.Itoa(os.Getpid())), 0644)
}

func runSession(configFile string, fileChangeDelay time.Duration, simpleStatus bool, c <-chan os.Signal) (int, bool) {
	monitor, err := buildFileMonitor(fileChangeDelay)
	if err != nil {
		return 2, false
	}

	var tree = config.BuildProcessTree(configFile, monitor)

	done := make(chan bool)

	statusChartQuit := statuschart.Start(tree, done, simpleStatus)
	clientHandlerQuit := clienthandler.Start(tree, done)
	slaveMonitorQuit := processtree.StartSlaveMonitor(tree, monitor.Listen(), done)

	sig := <-c

	// Tear down in reverse startup order
	exit(slaveMonitorQuit, done)
	exit(clientHandlerQuit, done)
	monitor.Close()
	exit(statusChartQuit, done)

	if sig == syscall.SIGUSR1 {
		return 0, true
	}
	if sig == syscall.SIGINT {
		return 0, false
	}
	return 1, false
}

func exit(quit, done chan bool) {
	// Signal the process to quit.
	close(quit)
	// Wait until the process signals it's done.
	<-done
}

func buildFileMonitor(fileChangeDelay time.Duration) (filemonitor.FileMonitor, error) {
	if portStr := os.Getenv(listenerPortVar); portStr != "" {
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return nil, fmt.Errorf("%s must be an integer or empty string: %v", listenerPortVar, err)
		}

		ln, err := net.ListenTCP("tcp", &net.TCPAddr{
			IP:   net.ParseIP("127.0.0.1"),
			Port: port,
		})
		if err != nil {
			return nil, err
		}

		return filemonitor.NewFileListener(fileChangeDelay, ln), nil
	}

	monitor, err := filemonitor.NewFileMonitor(fileChangeDelay)
	if err != nil {
		return nil, err
	}

	return monitor, nil
}
