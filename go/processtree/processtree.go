package processtree

import (
	"fmt"
	"sync"

	"github.com/burke/zeus/go/unixsocket"
)

type ProcessTree struct {
	Root         *SlaveNode
	ExecCommand  string
	SlavesByName map[string]*SlaveNode
	Commands     []*CommandNode
	StateChanged chan bool

	restartMutex sync.Mutex
}

type ProcessTreeNode struct {
	mu     sync.RWMutex
	Parent *SlaveNode
	Name   string
}

type CommandNode struct {
	ProcessTreeNode
	booting sync.RWMutex
	Aliases []string
}

func (tree *ProcessTree) Start(fileChanges <-chan []string, done chan bool) chan bool {
	quit := make(chan bool)
	go func() {
		for _, slave := range tree.SlavesByName {
			go slave.Run(tree.StateChanged, tree.ExecCommand)
			defer slave.ForceKill()
		}

		for {
			select {
			case <-quit:
				done <- true
				return
			case files := <-fileChanges:
				if len(files) > 0 {
					tree.RestartNodesWithFeatures(files)
				}
			}
		}
	}()
	return quit
}

func (tree *ProcessTree) NewCommandNode(name string, aliases []string, parent *SlaveNode) *CommandNode {
	x := &CommandNode{}
	x.Parent = parent
	x.Name = name
	x.Aliases = aliases
	tree.Commands = append(tree.Commands, x)
	return x
}

func (tree *ProcessTree) BootCommand(requested string) (*unixsocket.Usock, string, error) {
	cmdNode := tree.findCommand(requested)
	if cmdNode == nil {
		return nil, "", fmt.Errorf("unknown command %s", requested)
	}

	sock, err := cmdNode.Parent.BootCommand(requested)
	if err != nil {
		return nil, "", err
	}

	return sock, cmdNode.Name, nil
}

func (tree *ProcessTree) findCommand(requested string) *CommandNode {
	for _, command := range tree.Commands {
		if command.Name == requested {
			return command
		}
		for _, alias := range command.Aliases {
			if alias == requested {
				return command
			}
		}
	}
	return nil
}

func (tree *ProcessTree) AllCommandsAndAliases() []string {
	var values []string
	for _, command := range tree.Commands {
		values = append(values, command.Name)
		for _, alias := range command.Aliases {
			values = append(values, alias)
		}
	}
	return values
}

func (tree *ProcessTree) RestartNodesWithFeatures(files []string) {
	tree.restartMutex.Lock()
	defer tree.restartMutex.Unlock()
	tree.Root.trace("%d files changed, beginning with %q", len(files), files[0])
	tree.Root.restartNodesWithFeatures(tree, files)
}

// Serialized: restartMutex is always held when this is called.
func (node *SlaveNode) restartNodesWithFeatures(tree *ProcessTree, files []string) {
	for _, file := range files {
		if node.HasFeature(file) {
			node.trace("restarting for %q", file)
			node.RequestRestart()
			return
		}
	}
	for _, s := range node.Slaves {
		s.restartNodesWithFeatures(tree, files)
	}
}
