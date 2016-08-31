package statuschart

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/burke/ttyutils"
	"github.com/burke/zeus/go/processtree/node"
	slog "github.com/burke/zeus/go/shinylog"

	"os"

	"github.com/burke/zeus/go/processtree"
)

const (
	lineT = "{yellow}├── "
	lineL = "{yellow}└── "
	lineI = "{yellow}│   "
	lineX = "{yellow}    "
)

type ttyChart struct {
	output       *os.File
	termios      *ttyutils.Termios
	log          *slog.ShinyLogger
	extraOutput  bytes.Buffer
	state        processtree.State
	drawnInitial bool
}

func newTTYChart(output *os.File) (*ttyChart, error) {
	if !ttyutils.IsTerminal(output.Fd()) {
		return nil, errors.New("output is not a terminal")
	}

	termios, err := ttyutils.NoEcho(uintptr(output.Fd()))
	if err != nil {
		return nil, err
	}

	chart := ttyChart{
		output:  output,
		termios: termios,
		log:     slog.NewShinyLogger(output, output),
	}

	return &chart, nil
}

func (t *ttyChart) watchOutput(output *os.File) {
	scw := &StringChannelWriter{make(chan string, 10)}
	slog.SetDefaultLogger(slog.NewShinyLogger(scw, scw))

	for out := range scw.Notif {
		if t.drawnInitial {
			fmt.Printf(out)
		}
		t.extraOutput.WriteString(out)
		t.draw()
	}
}

func (t *ttyChart) Update(state processtree.State) {
	t.state = state
	t.draw()
}

func (t *ttyChart) Stop() error {
	return ttyutils.RestoreTerminalState(uintptr(t.output.Fd()), t.termios)
}

func (t *ttyChart) draw() {
	extraStr := t.extraOutput.String()

	if t.drawnInitial {
		numberOfOutputLines := len(t.state.Commands) + len(t.state.NodeState) +
			t.stringOutputLines(extraStr) + 3
		fmt.Printf("\033[%dA", numberOfOutputLines)
	} else {
		t.drawnInitial = true
	}

	t.log.Colorized("\x1b[4m{green}[ready] {red}[crashed] {blue}[running] {magenta}[connecting] {yellow}[waiting]\033[K")
	for _, root := range t.state.RootNodes {
		t.drawSubtree(root, "", "")
	}

	t.log.Colorized("\033[K\n\x1b[4mAvailable Commands: {yellow}[waiting] {red}[crashed] {green}[ready]\033[K")
	t.drawCommands()
	output := strings.Replace(extraStr, "\n", "\033[K\n", -1)
	fmt.Printf(output)
}

func (t *ttyChart) stringOutputLines(out string) int {
	ts, err := ttyutils.Winsize(t.output)
	if err != nil {
		// This can happen when the output is redirected to a device
		// that blows up on the ioctl Winsize uses. We don't care about fancy drawing in this case.
		return 0
	}
	width := int(ts.Columns)
	if width == 0 { // output has been redirected
		return 0
	}

	lines := strings.Split(out, "\n")

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

func (t *ttyChart) drawCommands() {
	commandAliases := collectCommandAliases(t.state)

	for cmd, parent := range t.state.Commands {
		alia := strings.Join(commandAliases[cmd], ", ")
		var aliasPart string
		if len(alia) > 0 {
			aliasPart = " (alias: " + alia + ")"
		}
		text := "zeus " + cmd + aliasPart
		reset := "\033[K"

		switch t.state.NodeState[parent] {
		case node.SReady:
			t.log.Green(text + reset)
		case node.SCrashed:
			t.log.Red(text + " {yellow}[run to see backtrace]" + reset)
		default:
			t.log.Yellow(text + reset)
		}
	}
}

func (t *ttyChart) drawSubtree(name, myIndentation, childIndentation string) {
	printStateInfo(t.log, myIndentation, name, t.state.NodeState[name], false, true)

	children := t.state.NodeTree[name]
	for i, child := range children {
		if i == len(children)-1 {
			t.drawSubtree(child, childIndentation+lineL, childIndentation+lineX)
		} else {
			t.drawSubtree(child, childIndentation+lineT, childIndentation+lineI)
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
