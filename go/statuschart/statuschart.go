package statuschart

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/burke/ttyutils"
	"github.com/burke/zeus/go/processtree"
	slog "github.com/burke/zeus/go/shinylog"
)

const updateDebounceInterval = 1 * time.Millisecond

type StatusChart struct {
	RootSlave *processtree.SlaveNode
	update    chan bool

	numberOfSlaves int
	Commands       []*processtree.CommandNode
	L              sync.Mutex
	drawnInitial   bool

	directLogger *slog.ShinyLogger

	extraOutput       string
	terminalSupported bool

	previousStates []*string
}

var theChart *StatusChart

func Start(tree *processtree.ProcessTree, done chan bool, simple bool) chan bool {
	quit := make(chan bool)

	theChart = &StatusChart{}
	theChart.RootSlave = tree.Root
	theChart.numberOfSlaves = len(tree.SlavesByName)
	theChart.Commands = tree.Commands
	theChart.update = make(chan bool)
	theChart.directLogger = slog.NewShinyLogger(os.Stdout, os.Stderr)
	theChart.terminalSupported = ttyutils.IsTerminal(os.Stdout.Fd())

	if simple {
		startLineOutput(tree, done, quit)
	} else {
		if theChart.terminalSupported {
			ttyStart(tree, done, quit)
		} else {
			stdoutStart(tree, done, quit)
		}
	}

	go theChart.watchUpdates(tree.StateChanged)

	return quit
}

func startLineOutput(tree *processtree.ProcessTree, done, quit chan bool) {
	states := make(map[string]string)

	go func() {
		for {
			select {
			case <-quit:
				done <- true
				return
			case <-theChart.update:
				for name, slave := range tree.SlavesByName {
					state, found := states[name]
					if !found || (state != slave.State()) {
						fmt.Println("environment: " + name + " status: " + slave.HumanReadableState())
						states[name] = slave.State()
					}
				}
			}
		}
	}()
}

func (s *StatusChart) watchUpdates(updates <-chan bool) {
	// Debounce state updates
	for <-updates {
		reported := false
		timeout := time.After(updateDebounceInterval)
		for !reported {
			select {
			case <-updates:
			case <-timeout:
				s.update <- true
				reported = true
			}
		}
	}
}

func stateSuffix(state string) string {
	status := ""

	switch state {
	case processtree.SUnbooted:
		status = "{U}"
	case processtree.SBooting:
		status = "{B}"
	case processtree.SCrashed:
		status = "{!C}"
	case processtree.SReady:
		status = "{R}"
	default:
		status = "{?}"
	}

	return status
}

func printStateInfo(indentation, identifier, state string, verbose, printNewline bool) {
	log := theChart.directLogger
	newline := ""
	suffix := ""
	if printNewline {
		newline = "\n"
	}
	if verbose {
		suffix = stateSuffix(state)
	}
	switch state {
	case processtree.SUnbooted:
		log.ColorizedSansNl(indentation + "{magenta}" + identifier + suffix + "\033[K" + newline)
	case processtree.SBooting:
		log.ColorizedSansNl(indentation + "{blue}" + identifier + suffix + "\033[K" + newline)
	case processtree.SCrashed:
		log.ColorizedSansNl(indentation + "{red}" + identifier + suffix + "\033[K" + newline)
	case processtree.SReady:
		// no status suffix, as that's the optimal state
		log.ColorizedSansNl(indentation + "{green}" + identifier + suffix + "\033[K" + newline)
	default:
		log.ColorizedSansNl(indentation + "{yellow}" + identifier + suffix + "\033[K" + newline)
	}
}
