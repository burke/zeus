package zeusmaster

import (
	"net"
	"sync"
)

type ProcessTree struct {
	Root *SlaveNode
	ExecCommand string
	slavesByName map[string]*SlaveNode
	commandsByName map[string]*CommandNode
}

type ProcessTreeNode struct {
	mu sync.RWMutex
	Parent *SlaveNode
	Name string
}

type SlaveNode struct {
	ProcessTreeNode
	Socket *net.UnixConn
	Pid int
	Slaves []*SlaveNode
	Commands []*CommandNode
	Features map[string]bool
}

type CommandNode struct {
	ProcessTreeNode
	Aliases []string
}

func (tree *ProcessTree) NewCommandNode(name string, aliases []string, parent *SlaveNode) *CommandNode {
	x := &CommandNode{}
	x.Parent = parent
	x.Name = name
	tree.commandsByName[name] = x
	return x
}

func (tree *ProcessTree) NewSlaveNode(name string, parent *SlaveNode) *SlaveNode {
	x := &SlaveNode{}
	x.Parent = parent
	x.Name = name
	tree.slavesByName[name] = x
	return x
}

func (tree *ProcessTree) FindSlaveByName(name string) *SlaveNode {
	if name == "" {
		return tree.Root
	}
	return tree.slavesByName[name]
}

func (tree *ProcessTree) FindCommandByName(name string) *CommandNode {
	return tree.commandsByName[name]
}
