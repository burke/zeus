package zeusmaster

import (
	"sync"
)

type ProcessTree struct {
	Root *SlaveNode
	ExecCommand string
	SlavesByName map[string]*SlaveNode
	CommandsByName map[string]*CommandNode
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
	tree.CommandsByName[name] = x
	return x
}

func (tree *ProcessTree) NewSlaveNode(name string, parent *SlaveNode) *SlaveNode {
	x := &SlaveNode{}
	x.Parent = parent
	x.SignalUnbooted()
	x.Name = name
	x.ClientCommandPTYFileDescriptor = make(chan int)
	tree.SlavesByName[name] = x
	return x
}

func (tree *ProcessTree) FindSlaveByName(name string) *SlaveNode {
	if name == "" {
	return tree.Root
	}
	return tree.SlavesByName[name]
}

func (tree *ProcessTree) FindCommandByName(name string) *CommandNode {
	return tree.CommandsByName[name]
}

func (tree *ProcessTree) AllCommandsAndAliases() []string {
	var values []string
	for name, command := range tree.CommandsByName {
		values = append(values, name)
		for _, alias := range command.Aliases {
			values = append(values, alias)
		}
	}
	return values
}

func (tree *ProcessTree) KillNodesWithFeature(file string) {
	tree.Root.killNodesWithFeature(tree, file)
}

func (node *SlaveNode) killNodesWithFeature(tree *ProcessTree, file string) {
	if node.Features[file] {
		node.Kill(tree, )
	} else {
		for _, s := range node.Slaves {
			s.killNodesWithFeature(tree, file)
		}
	}
}