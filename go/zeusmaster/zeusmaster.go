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
var terminatingSignals = []os.Signal{syscall.SIGHUP, syscall.SIGINT, syscall.SIGKILL, syscall.SIGALRM, syscall.SIGTERM, syscall.SIGXCPU, syscall.SIGXFSZ, syscall.SIGVTALRM, syscall.SIGPROF, syscall.SIGUSR1, syscall.SIGUSR2}

func Run(configFile string, fileChangeDelay time.Duration, simpleStatus bool) int {
	slog.Colorized("{green}Starting {yellow}Z{red}e{blue}u{magenta}s{green} server v" + zeusversion.VERSION)

	zerror.Init()

	monitor, err := buildFileMonitor(fileChangeDelay)
	if err != nil {
		return 2
	}

	var tree = config.BuildProcessTree(configFile, monitor)

	done := make(chan bool)

	defer exit(processtree.StartSlaveMonitor(tree, monitor.Listen(), done), done)
	defer exit(clienthandler.Start(tree, done), done)
	defer monitor.Close()
	defer slog.Suppress()
	defer zerror.PrintFinalOutput()
	defer exit(statuschart.Start(tree, done, simpleStatus), done)

	c := make(chan os.Signal, 1)
	signal.Notify(c, terminatingSignals...)

	for {
		select {
		case sig := <-c:
			if sig == syscall.SIGINT {
				return 0
			}
			return 1
		}
	}
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
