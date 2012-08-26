package zeusmaster

type ProcessTree struct {
	Root *SlaveNode
	ExecCommand string
	nodesByName map[string]*ProcessTreeNode
}

type ProcessTreeNode struct {
	Name string
	Action string
}

type SlaveNode struct {
	ProcessTreeNode
	Pid int
	Slaves []*SlaveNode
	Commands []*CommandNode
	Features map[string]bool
}

type CommandNode struct {
	ProcessTreeNode
	Aliases []string
}

func NewCommandNode(tree *ProcessTree, name string, aliases []string) (*CommandNode) {
	x := CommandNode{}
	x.Name = name
	tree.nodesByName[name] = &x.ProcessTreeNode
	return &x
}

func NewSlaveNode(tree *ProcessTree, name string) (*SlaveNode) {
	x := SlaveNode{}
	x.Name = name
	tree.nodesByName[name] = &x.ProcessTreeNode
	return &x
}

func (tree *ProcessTree) FindNodeByName(name string) *ProcessTreeNode {
	return tree.nodesByName[name]
}
