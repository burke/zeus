package statuschart

import (
	"fmt"
	"github.com/burke/ttyutils"
	slog "github.com/burke/zeus/go/shinylog"
	"strings"

	"github.com/burke/zeus/go/processtree"
	"os"
)

const (
	lineT = "{yellow}├── "
	lineL = "{yellow}└── "
	lineI = "{yellow}│   "
	lineX = "{yellow}    "
)

func ttyStart(tree *processtree.ProcessTree, done, quit chan bool) {
	go func() {
		scw := &StringChannelWriter{make(chan string, 10)}
		slog.SetDefaultLogger(slog.NewShinyLogger(scw, scw))

		termios, err := ttyutils.NoEcho(uintptr(os.Stdout.Fd()))
		if err != nil {
			theChart.terminalSupported = false
		}

		for {
			select {
			case <-quit:
				ttyutils.RestoreTerminalState(uintptr(os.Stdout.Fd()), termios)
				done <- true
				return
			case output := <-scw.Notif:
				theChart.L.Lock()
				if theChart.drawnInitial {
					print(output)
				}
				theChart.extraOutput += output
				theChart.L.Unlock()
				theChart.draw()
			case <-tree.StateChanged:
				theChart.draw()
			case <-theChart.update:
				theChart.draw()
			}
		}
	}()
}

func (s *StatusChart) draw() {
	s.L.Lock()
	defer s.L.Unlock()

	if s.drawnInitial {
		lengthOfOutput := s.lengthOfOutput()
		numberOfOutputLines := s.numberOfSlaves + len(s.Commands) + lengthOfOutput + 3
		fmt.Printf("\033[%dA", numberOfOutputLines)
	} else {
		s.drawnInitial = true
	}

	log := theChart.directLogger

	log.Colorized("\x1b[4m{green}[ready] {red}[crashed] {blue}[running] {magenta}[connecting] {yellow}[waiting]\033[K")
	s.drawSubtree(s.RootSlave, "", "")

	log.Colorized("\033[K\n\x1b[4mAvailable Commands: {yellow}[waiting] {red}[crashed] {green}[ready]\033[K")
	s.drawCommands()
	output := strings.Replace(s.extraOutput, "\n", "\033[K\n", -1)
	fmt.Printf(output)
}

func (s *StatusChart) lengthOfOutput() int {
	ts, err := ttyutils.Winsize(os.Stdout)
	if err != nil {
		// This can happen when the output is redirected to a device
		// that blows up on the ioctl Winsize uses. We don't care about fancy drawing in this case.
		return 0
	}
	width := int(ts.Columns)
	if width == 0 { // output has been redirected
		return 0
	}

	lines := strings.Split(s.extraOutput, "\n")

	numLines := 0
	for _, line := range lines {
		n := (len(line) + width - 1) / width
		if n == 0 {
			n = 1
		}
		numLines += n
	}

	return numLines - 1
}

func (s *StatusChart) drawCommands() {
	for _, command := range s.Commands {
		state := command.Parent.State

		alia := strings.Join(command.Aliases, ", ")
		var aliasPart string
		if len(alia) > 0 {
			aliasPart = " (alias: " + alia + ")"
		}
		text := "zeus " + command.Name + aliasPart
		reset := "\033[K"

		log := theChart.directLogger

		switch state {
		case processtree.SReady:
			log.Green(text + reset)
		case processtree.SCrashed:
			log.Red(text + " {yellow}[run to see backtrace]" + reset)
		default:
			log.Yellow(text + reset)
		}
	}
}

func (s *StatusChart) drawSubtree(node *processtree.SlaveNode, myIndentation, childIndentation string) {
	printStateInfo(myIndentation, node.Name, node.State, false, true)

	for i, slave := range node.Slaves {
		if i == len(node.Slaves)-1 {
			s.drawSubtree(slave, childIndentation+lineL, childIndentation+lineX)
		} else {
			s.drawSubtree(slave, childIndentation+lineT, childIndentation+lineI)
		}
	}
}

type StringChannelWriter struct {
	Notif chan string
}

func (s *StringChannelWriter) Write(o []byte) (int, error) {
	s.Notif <- string(o)
	return len(o), nil
}
