package processtree

import (
	"fmt"
	"sync"

	"github.com/burke/zeus/go/filemonitor"
	"github.com/burke/zeus/go/processtree/node"
	"github.com/burke/zeus/go/processtree/process"
)

func NewProcessTree(monitor filemonitor.FileMonitor) ProcessTree {
	return &tree{
		monitor: monitor,

		nodes:     make(map[string]*node.Node),
		commands:  make(map[string]string),
		aliases:   make(map[string]string),
		nodeTree:  make(map[string][]string),
		nodeState: make(map[string]node.State),
		watches:   make(map[string]map[string]struct{}),
	}
}

type ProcessTree interface {
	Run(chan<- State)
	Stop()
	AddChildNode(name, parent string) error
	AddRootNode(name string, command []string) error
	AddCommand(name, parent string, aliases []string) error
	BootCommand(name string, client process.CommandClient) (process.CommandProcess, error)
}

type State struct {
	NodeState map[string]node.State
	Commands  map[string]string
	Aliases   map[string]string
	NodeTree  map[string][]string
	RootNodes []string
}

type tree struct {
	nodes       map[string]*node.Node // node name to node instance
	commands    map[string]string     // command name to node name
	aliases     map[string]string     // alias name to command name
	nodeTree    map[string][]string   // node name to child node names
	rootNodes   []string              // Slice of root node names
	rootCommand []string              // Command line arguments for starting the root command

	nodeState      map[string]node.State // node name to node state
	nodeStateMutex sync.Mutex

	monitor    filemonitor.FileMonitor
	watches    map[string]map[string]struct{} // node name to set of files watched
	watchMutex sync.Mutex

	running sync.WaitGroup
}

func (t *tree) Run(state chan<- State) {
	go t.reloadOnChanges(t.monitor.Listen())

	for name, n := range t.nodes {
		t.running.Add(1)
		go func(name string, n *node.Node) {
			nodeState := make(chan node.State)
			go n.Run(nodeState)

			for s := range nodeState {
				t.updateState(name, s)
			}
			t.running.Done()
		}(name, n)
	}
	go func() {
		t.running.Wait()
		close(state)
	}()
}

func (t *tree) Stop() {
	for _, n := range t.nodes {
		go func(n *node.Node) {
			n.Stop()
		}(n)
	}
	t.running.Wait()
}

func (t *tree) AddRootNode(name string, command []string) error {
	t.addNode(name, func() (process.NodeProcess, error) {
		return process.StartProcess(command)
	})
	t.rootNodes = append(t.rootNodes, name)

	return nil
}

func (t *tree) AddChildNode(name, parent string) error {
	parentNode, ok := t.nodes[parent]
	if !ok {
		return fmt.Errorf("Unknown parent node %q for new node %q", parent, name)
	}

	t.addNode(name, func() (process.NodeProcess, error) {
		return parentNode.BootNode(name)
	})
	t.nodeTree[parent] = append(t.nodeTree[parent], name)

	return nil
}

func (t *tree) addNode(name string, boot func() (process.NodeProcess, error)) {
	t.nodes[name] = node.NewNode(name, boot, func(file string) {
		t.watchFile(name, file)
	})
	t.nodeState[name] = node.SUnbooted
	t.watches[name] = make(map[string]struct{})
	t.nodeTree[name] = nil
}

func (t *tree) AddCommand(cmd, parent string, aliases []string) error {
	if _, ok := t.nodes[parent]; !ok {
		return fmt.Errorf("Unknown parent node %q for command %q", parent, cmd)
	}

	t.commands[cmd] = parent
	for _, alias := range aliases {
		t.aliases[alias] = cmd
	}

	return nil
}

func (t *tree) BootCommand(name string, client process.CommandClient) (process.CommandProcess, error) {
	if aliased, ok := t.aliases[name]; ok {
		name = aliased
	}

	nodeName, ok := t.commands[name]
	if !ok {
		return nil, fmt.Errorf("Unknown command %q", name)
	}

	parentNode := t.nodes[nodeName]
	return parentNode.BootCommand(name, client)
}

func (t *tree) watchFile(node, file string) {
	t.watchMutex.Lock()
	defer t.watchMutex.Unlock()

	t.watches[node][file] = struct{}{}
	t.monitor.Add(file)
}

func (t *tree) reloadOnChanges(listener <-chan []string) {
	for changes := range listener {
		t.watchMutex.Lock()
		shouldReload := make(map[string]struct{})
		for nodeName, nodeWatches := range t.watches {
			for _, file := range changes {
				if _, ok := nodeWatches[file]; ok {
					shouldReload[nodeName] = struct{}{}
					break
				}
			}
		}

		for nodeName := range shouldReload {
			for _, child := range t.allChildren(nodeName) {
				shouldReload[child] = struct{}{}
			}
		}

		for nodeName := range shouldReload {
			t.nodes[nodeName].FileChanged()
		}

		t.watchMutex.Unlock()
	}
}

func (t *tree) allChildren(nodeName string) []string {
	children := t.nodeTree[nodeName]
	for _, child := range children {
		children = append(children, t.allChildren(child)...)
	}

	return children
}

func (t *tree) updateState(nodeName string, nodeState node.State) State {
	t.nodeStateMutex.Lock()
	defer t.nodeStateMutex.Unlock()

	t.nodeState[nodeName] = nodeState

	// Regenerating the state object on every change isn't exactly
	// efficient but these trees should be small enough that it
	// isn't a performance bottleneck.
	state := State{
		NodeState: make(map[string]node.State, len(t.nodeState)),
		Commands:  make(map[string]string, len(t.commands)),
		Aliases:   make(map[string]string, len(t.aliases)),
		NodeTree:  make(map[string][]string, len(t.nodeTree)),
		RootNodes: make([]string, len(t.rootNodes)),
	}

	for k, v := range t.nodeState {
		state.NodeState[k] = v
	}
	for k, v := range t.commands {
		state.Commands[k] = v
	}
	for k, v := range t.aliases {
		state.Aliases[k] = v
	}
	for k, v := range t.nodeTree {
		nv := make([]string, len(v))
		copy(nv, v)
		state.NodeTree[k] = nv
	}
	copy(state.RootNodes, t.rootNodes)

	return state
}
