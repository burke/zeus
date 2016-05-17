package processtree

import (
	"sync"
)

type ProcessTree struct {
	Root         *SlaveNode
	ExecCommand  string
	SlavesByName map[string]*SlaveNode
	Commands     []*CommandNode
	StateChanged chan bool
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

func (tree *ProcessTree) NewCommandNode(name string, aliases []string, parent *SlaveNode) *CommandNode {
	x := &CommandNode{}
	x.Parent = parent
	x.Name = name
	x.Aliases = aliases
	tree.Commands = append(tree.Commands, x)
	return x
}

func (tree *ProcessTree) FindSlaveByName(name string) *SlaveNode {
	if name == "" {
		return tree.Root
	}
	return tree.SlavesByName[name]
}

func (tree *ProcessTree) FindCommand(requested string) *CommandNode {
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

var restartMutex sync.Mutex

func (tree *ProcessTree) RestartNodesWithFeatures(files []string) {
	restartMutex.Lock()
	defer restartMutex.Unlock()
	tree.Root.restartNodesWithFeatures(tree, files)
}

// Serialized: restartMutex is always held when this is called.
func (node *SlaveNode) restartNodesWithFeatures(tree *ProcessTree, files []string) {
	for _, file := range files {
		if node.Features[file] {
			node.RequestRestart()
			return
		}
	}
	for _, s := range node.Slaves {
		s.restartNodesWithFeatures(tree, files)
	}
}
