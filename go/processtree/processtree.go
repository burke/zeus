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
		values = append(values, command.Aliases...)
	}
	return values
}

var restartMutex sync.Mutex

func (tree *ProcessTree) RestartNodesWithFeatures(files []string) {
	restartMutex.Lock()
	defer restartMutex.Unlock()
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

// We implement sort.Interface - Len, Less, and Swap - on list of commands so
// we can use the sort package’s generic Sort function.
type Commands []*CommandNode

func (c Commands) Len() int {
	return len(c)
}

func (c Commands) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

func (c Commands) Less(i, j int) bool {
	return c[i].Name < c[j].Name
}
