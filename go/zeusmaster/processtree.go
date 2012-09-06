package zeusmaster

import (
	"sync"
)

type ProcessTree struct {
	Root *SlaveNode
	ExecCommand string
	SlavesByName map[string]*SlaveNode
	Commands []*CommandNode
}

type ProcessTreeNode struct {
	mu sync.RWMutex
	Parent *SlaveNode
	Name string
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

func (tree *ProcessTree) RestartNodesWithFeature(file string) {
	tree.Root.restartNodesWithFeature(tree, file)
}

func (node *SlaveNode) restartNodesWithFeature(tree *ProcessTree, file string) {
	if node.Features[file] {
		node.RequestRestart()
	} else {
		for _, s := range node.Slaves {
			s.restartNodesWithFeature(tree, file)
		}
	}
}