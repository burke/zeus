package zeusmaster

import (
	"fmt"
	"github.com/burke/ttyutils"
	slog "github.com/burke/zeus/go/shinylog"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	lineT = "\x1b[33m├── "
	lineL = "\x1b[33m└── "
	lineI = "\x1b[33m│   "
	lineX = "\x1b[33m    "
)

type StatusChart struct {
	RootSlave *SlaveNode
	update    chan bool

	numberOfSlaves int
	Commands       []*CommandNode
	L              sync.Mutex
	drawnInitial   bool
}

var theChart *StatusChart

func StartStatusChart(tree *ProcessTree, quit chan bool) {
	theChart = &StatusChart{}
	theChart.RootSlave = tree.Root
	theChart.numberOfSlaves = len(tree.SlavesByName)
	theChart.Commands = tree.Commands
	theChart.update = make(chan bool, 10)

	termios, err := ttyutils.NoEcho(uintptr(os.Stdout.Fd()))
	if err != nil {
		slog.Error(err)
	}

	ticker := time.Tick(1000 * time.Millisecond)

	for {
		select {
		case <-quit:
			ttyutils.RestoreTerminalState(uintptr(os.Stdout.Fd()), termios)
			quit <- true
			return
		case <-ticker:
			theChart.draw()
		case <-theChart.update:
			theChart.draw()
		}
	}
}

func StatusChartUpdate() {
	theChart.update <- true
}

func printStateInfo(indentation, identifier, state string, verbose bool) {
	switch state {
	case sUnbooted:
		if verbose {
			slog.Colorized(indentation + "{magenta}" + identifier)
		}
	case sBooting:
		slog.Colorized(indentation + "{blue}" + identifier)
	case sCrashed:
		slog.Colorized(indentation + "{red}" + identifier)
	case sReady:
		slog.Colorized(indentation + "{green}" + identifier)
	case sWaiting:
		fallthrough
	default:
		slog.Colorized(indentation + "{yellow}" + identifier)
	}
}

func (s *StatusChart) draw() {
	s.L.Lock()
	defer s.L.Unlock()

	if s.drawnInitial {
		fmt.Printf("\033[%dA", s.numberOfSlaves+3+len(s.Commands))
	} else {
		s.drawnInitial = true
	}

	slog.Colorized("\x1b[4m{green}[ready] {red}[crashed] {blue}[running] {magenta}[connecting] {yellow}[waiting]")
	s.drawSubtree(s.RootSlave, "", "")

	slog.Colorized("\n\x1b[4mAvailable Commands: {yellow}[waiting] {red}[crashed] {green}[ready]")
	s.drawCommands()
}

func (s *StatusChart) drawCommands() {
	for _, command := range s.Commands {
		state := command.Parent.state

		alia := strings.Join(command.Aliases, ", ")
		var aliasPart string
		if len(alia) > 0 {
			aliasPart = " (alias: " + alia + ")"
		}
		text := "zeus " + command.Name + aliasPart

		switch state {
		case sReady:
			slog.Green(text)
		case sCrashed:
			slog.Red(text)
		default:
			slog.Yellow(text)
		}
	}
}

func (s *StatusChart) drawSubtree(node *SlaveNode, myIndentation, childIndentation string) {
	printStateInfo(myIndentation, node.Name, node.state, true)

	for i, slave := range node.Slaves {
		if i == len(node.Slaves)-1 {
			s.drawSubtree(slave, childIndentation+lineL, childIndentation+lineX)
		} else {
			s.drawSubtree(slave, childIndentation+lineT, childIndentation+lineI)
		}
	}
}
