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
	lineT = "{yellow}├── "
	lineL = "{yellow}└── "
	lineI = "{yellow}│   "
	lineX = "{yellow}    "
)

type StatusChart struct {
	RootSlave *SlaveNode
	update    chan bool

	numberOfSlaves int
	Commands       []*CommandNode
	L              sync.Mutex
	drawnInitial   bool

	directLogger *slog.ShinyLogger

	extraOutput string
}

var theChart *StatusChart

func StartStatusChart(tree *ProcessTree, done chan bool) chan bool {
	quit := make(chan bool)
	go func() {
		theChart = &StatusChart{}
		theChart.RootSlave = tree.Root
		theChart.numberOfSlaves = len(tree.SlavesByName)
		theChart.Commands = tree.Commands
		theChart.update = make(chan bool, 10)
		theChart.directLogger = slog.NewShinyLogger(os.Stdout, os.Stderr)

		scw := &StringChannelWriter{make(chan string, 10)}
		slog.DefaultLogger = slog.NewShinyLogger(scw, scw)

		termios, err := ttyutils.NoEcho(uintptr(os.Stdout.Fd()))
		if err != nil {
			slog.Error(err)
		}

		ticker := time.Tick(1000 * time.Millisecond)
		for {
			select {
			case <-quit:
				ttyutils.RestoreTerminalState(uintptr(os.Stdout.Fd()), termios)
				done <- true
				return
			case <-ticker:
				theChart.draw()
			case output := <-scw.Notif:
				theChart.L.Lock()
				if theChart.drawnInitial {
					print(output)
				}
				theChart.extraOutput += output
				theChart.L.Unlock()
				theChart.draw()
			case <-theChart.update:
				theChart.draw()
			}
		}
	}()
	return quit
}

func StatusChartUpdate() {
	theChart.update <- true
}

func printStateInfo(indentation, identifier, state string, verbose bool) {
	log := theChart.directLogger
	switch state {
	case sUnbooted:
		if verbose {
			log.Colorized(indentation + "{magenta}" + identifier + "\033[K")
		}
	case sBooting:
		log.Colorized(indentation + "{blue}" + identifier + "\033[K")
	case sCrashed:
		log.Colorized(indentation + "{red}" + identifier + "\033[K")
	case sReady:
		log.Colorized(indentation + "{green}" + identifier + "\033[K")
	case sWaiting:
		fallthrough
	default:
		log.Colorized(indentation + "{yellow}" + identifier + "\033[K")
	}
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
		state := command.Parent.state

		alia := strings.Join(command.Aliases, ", ")
		var aliasPart string
		if len(alia) > 0 {
			aliasPart = " (alias: " + alia + ")"
		}
		text := "zeus " + command.Name + aliasPart + "\033[K"

		log := theChart.directLogger

		switch state {
		case sReady:
			log.Green(text)
		case sCrashed:
			log.Red(text)
		default:
			log.Yellow(text)
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

type StringChannelWriter struct {
	Notif chan string
}

func (s *StringChannelWriter) Write(o []byte) (int, error) {
	s.Notif <- string(o)
	return len(o), nil
}
