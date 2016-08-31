package statuschart

import (
	"strings"

	"github.com/burke/zeus/go/processtree"
	"github.com/burke/zeus/go/processtree/node"
	slog "github.com/burke/zeus/go/shinylog"
)

type stdoutChart struct {
	log *slog.ShinyLogger
}

func (s *stdoutChart) Update(state processtree.State) {
	s.log.ColorizedSansNl("{reset}Status: ")
	for _, root := range state.RootNodes {
		s.logSubtree(state, root)
	}
	s.log.Colorized("{reset}")
	s.logCommands(state)
}

func (s *stdoutChart) Stop() error {
	return nil
}

func collectCommands(state processtree.State, desiredNodeState node.State) []string {
	var commands []string

	for cmd, parent := range state.Commands {
		if state.NodeState[parent] == desiredNodeState {
			commands = append(commands, cmd)
		}
	}

	return commands
}

func (s *stdoutChart) logCommands(state processtree.State) {
	commandAliases := collectCommandAliases(state)
	runningCommands := collectCommands(state, node.SReady)
	crashedCommands := collectCommands(state, node.SCrashed)

	if len(runningCommands) > 0 {
		s.log.ColorizedSansNl("Available commands: ")
		first := true
		for _, cmd := range runningCommands {
			if !first {
				s.log.ColorizedSansNl("{reset}, ")
			}
			first = false

			s.log.ColorizedSansNl("{green}" + cmd)
			aliases := commandAliases[cmd]
			if len(aliases) > 0 {
				s.log.ColorizedSansNl("{reset}(aliases: {green}" + strings.Join(aliases, ",{green}") +
					"{reset})")
			}
		}
		s.log.Colorized("{reset}")
	}

	if len(crashedCommands) > 0 {
		s.log.ColorizedSansNl("Crashed commands ({yellow}run to see backtrace{reset}): ")
		first := true
		for _, cmd := range crashedCommands {
			if !first {
				s.log.ColorizedSansNl("{reset}, ")
			}
			first = false

			s.log.ColorizedSansNl("{red}" + cmd)
		}
		s.log.Colorized("{reset}")
	}
}

func (s *stdoutChart) logSubtree(state processtree.State, name string) {
	printStateInfo(s.log, "", name, state.NodeState[name], true, false)

	children := state.NodeTree[name]
	if len(children) == 0 {
		return
	}

	s.log.ColorizedSansNl("{reset}(")
	for i, child := range children {
		if i > 0 {
			s.log.ColorizedSansNl("{reset}, ")
		}
		s.logSubtree(state, child)
	}
	s.log.ColorizedSansNl("{reset})")
}
