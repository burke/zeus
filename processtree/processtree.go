package processtree

type ProcessTree struct {
	Root SlaveNode
	ExecCommand string
}

type SlaveNode struct {
	Pid int
	Identifier string
	Action string
	Slaves []SlaveNode
	Commands []CommandNode
	Features map[string]bool
}

type CommandNode struct {
	Identifier string
	Aliases []string
	Action string
}
