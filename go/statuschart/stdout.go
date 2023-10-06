package statuschart

import (
	"strings"

	"github.com/burke/zeus/go/processtree"
)

func stdoutStart(tree *processtree.ProcessTree, done, quit chan bool) {
	go func() {
		for {
			select {
			case <-quit:
				done <- true
				return
				// TODO: Maybe handle SCW Write requests?
			case <-theChart.update:
				theChart.logChanges()
			}
		}
	}()
}

func (s *StatusChart) logChanges() {
	s.L.Lock()
	defer s.L.Unlock()
	log := theChart.directLogger

	log.ColorizedSansNl("{reset}Status: ")
	s.logSubtree(s.RootWorker)
	log.Colorized("{reset}")
	s.logCommands()
}

func collectCommands(commands []*processtree.CommandNode, desiredState string) []*processtree.CommandNode {
	desiredCommands := make([]*processtree.CommandNode, 0)

	for _, command := range commands {
		if command.Parent.State() == desiredState {
			desiredCommands = append(desiredCommands, command)
		}
	}
	return desiredCommands
}

func (s *StatusChart) logCommands() {
	log := theChart.directLogger

	runningCommands := collectCommands(s.Commands, processtree.SReady)
	crashedCommands := collectCommands(s.Commands, processtree.SCrashed)
	if len(runningCommands) > 0 {
		log.ColorizedSansNl("Available commands: ")
		for i, command := range runningCommands {
			if i > 0 {
				log.ColorizedSansNl("{reset}, ")
			}
			log.ColorizedSansNl("{green}" + command.Name)
			if len(command.Aliases) > 0 {
				log.ColorizedSansNl("{reset}(aliases: {green}" + strings.Join(command.Aliases, ",{green}") +
					"{reset})")
			}
		}
		log.Colorized("{reset}")
	}

	if len(crashedCommands) > 0 {
		log.ColorizedSansNl("Crashed commands ({yellow}run to see backtrace{reset}): ")
		for i, command := range crashedCommands {
			if i > 0 {
				log.ColorizedSansNl("{reset}, ")
			}
			log.ColorizedSansNl("{red}" + command.Name)
		}
		log.Colorized("{reset}")
	}
}

func (s *StatusChart) logSubtree(node *processtree.WorkerNode) {
	log := theChart.directLogger
	printStateInfo("", node.Name, node.State(), true, false)

	if len(node.Workers) > 0 {
		log.ColorizedSansNl("{reset}(")
	}
	for i, worker := range node.Workers {
		if i != 0 {
			log.ColorizedSansNl("{reset}, ")
		}
		s.logSubtree(worker)
	}
	if len(node.Workers) > 0 {
		log.ColorizedSansNl("{reset})")
	}
}
