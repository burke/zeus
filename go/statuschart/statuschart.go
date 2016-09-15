package statuschart

import (
	"os"
	"time"

	"github.com/burke/zeus/go/processtree"
	"github.com/burke/zeus/go/processtree/node"
	slog "github.com/burke/zeus/go/shinylog"
)

const updateDebounceInterval = 1 * time.Millisecond

type chart interface {
	Update(processtree.State)
	Stop() error
}

// Start begins outputting status updates to the provided file. When the update
// channel closes, it shuts down and closes the returned channel.
func Start(output *os.File, updates <-chan processtree.State) <-chan struct{} {
	stopped := make(chan struct{})

	var ch chart
	var err error
	ch, err = newTTYChart(output)
	if err != nil {
		ch = &stdoutChart{
			log: slog.NewShinyLogger(output, output),
		}
	}

	go func() {
		// Introduce latency in state updates to group events
		for status := range updates {
			reported := false
			timeout := time.After(updateDebounceInterval)
			for !reported {
				select {
				case status = <-updates:
				case <-timeout:
					ch.Update(status)
					reported = true
				}
			}
		}
		ch.Stop()
		close(stopped)
	}()

	return stopped
}

func stateSuffix(state node.State) string {
	status := ""

	switch state {
	case node.SUnbooted:
		status = "{U}"
	case node.SBooting:
		status = "{B}"
	case node.SCrashed:
		status = "{!C}"
	case node.SReady:
		status = "{R}"
	default:
		status = "{?}"
	}

	return status
}

func printStateInfo(log *slog.ShinyLogger, indentation, identifier string, state node.State, verbose, printNewline bool) {
	newline := ""
	suffix := ""
	if printNewline {
		newline = "\n"
	}
	if verbose {
		suffix = stateSuffix(state)
	}
	switch state {
	case node.SUnbooted:
		log.ColorizedSansNl(indentation + "{magenta}" + identifier + suffix + "\033[K" + newline)
	case node.SBooting:
		log.ColorizedSansNl(indentation + "{blue}" + identifier + suffix + "\033[K" + newline)
	case node.SCrashed:
		log.ColorizedSansNl(indentation + "{red}" + identifier + suffix + "\033[K" + newline)
	case node.SReady:
		// no status suffix, as that's the optimal state
		log.ColorizedSansNl(indentation + "{green}" + identifier + suffix + "\033[K" + newline)
	default:
		log.ColorizedSansNl(indentation + "{yellow}" + identifier + suffix + "\033[K" + newline)
	}
}

func collectCommandAliases(state processtree.State) map[string][]string {
	commands := make(map[string][]string, len(state.Commands))

	for alias, cmd := range state.Aliases {
		commands[cmd] = append(commands[cmd], alias)
	}

	return commands
}
