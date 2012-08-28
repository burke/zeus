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
	bootWait sync.RWMutex
	Slaves []*SlaveNode
	Commands []*CommandNode
	Features map[string]bool
}

type CommandNode struct {
	ProcessTreeNode
	booting sync.RWMutex
	Aliases []string
}

func (node *SlaveNode) WaitUntilBooted() {
	node.bootWait.RLock()
	node.bootWait.RUnlock()
}

func (node *SlaveNode) SignalBooted() {
	node.bootWait.Unlock()
}

func (node *SlaveNode) SignalUnbooted() {
	node.bootWait.Lock()
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
	x.SignalUnbooted()
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
