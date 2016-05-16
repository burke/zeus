package main

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"syscall"

	"time"

	"github.com/burke/zeus/go/config"
	"github.com/burke/zeus/go/restarter"
	slog "github.com/burke/zeus/go/shinylog"
	"github.com/burke/zeus/go/zeusclient"
	"github.com/burke/zeus/go/zeusmaster"
	"github.com/burke/zeus/go/zeusversion"
)

var color bool = true
var Args = os.Args[1:]

func main() {
	for Args = os.Args[1:]; Args != nil && len(Args) > 0 && Args[0][0] == '-'; Args = Args[1:] {
		switch Args[0] {
		case "--no-color":
			color = false
			slog.DisableColor()
		case "--log":
			tracefile, err := os.OpenFile(Args[1], os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
			if err == nil {
				slog.SetTraceLogger(slog.NewTraceLogger(tracefile))
				Args = Args[1:]
			} else {
				fmt.Printf("Could not open trace file %s\n", Args[1])
				return
			}
		case "--file-change-delay":
			if len(Args) > 1 {
				delay, err := time.ParseDuration(Args[1])
				if err != nil {
					execManPage("zeus")
				}
				Args = Args[1:]
				restarter.FileChangeWindow = delay
			} else {
				execManPage("zeus")
			}
		case "--config":
			_, err := os.Stat(Args[1])
			if err != nil {
				fmt.Printf("Config file doesn't exist: %s (%e)\n", Args[1], err)
				return
			}
			config.ConfigFile = Args[1]
			Args = Args[1:]
		case "--version":
			printVersion()
			return
		}
	}
	if len(Args) == 0 {
		execManPage("zeus")
	}

	// Don't confuse the master by sending *full* args to it; just those that are
	// not zeus-specific.
	config.Args = Args

	if generalHelpRequested(Args) {
		execManPage("zeus")
	} else if Args[0] == "help" {
		commandSpecificHelp(Args)
	} else if Args[0] == "version" {
		printVersion()
	} else if Args[0] == "start" {
		os.Exit(zeusmaster.Run())
	} else if Args[0] == "init" {
		zeusInit()
	} else if Args[0] == "commands" {
		zeusCommands()
	} else {
		tree := config.BuildProcessTree()
		for _, name := range tree.AllCommandsAndAliases() {
			if Args[0] == name {
				os.Exit(zeusclient.Run())
			}
		}

		commandNotFound(Args[0])
	}
}

func execManPage(page string) {
	binaryPath := os.Args[0]
	gemDir := path.Dir(path.Dir(binaryPath))
	manDir := path.Join(gemDir, "man/build")
	zeus := path.Join(manDir, page)
	syscall.Exec("/usr/bin/env", []string{"/usr/bin/env", "man", zeus}, os.Environ())
}

func red() string {
	if color {
		return "\x1b[31m"
	}
	return ""
}

func reset() string {
	if color {
		return "\x1b[0m"
	}
	return ""
}

func copyFile(from, to string) (err error) {
	var src, dst *os.File
	wd, _ := os.Getwd()
	target := path.Join(wd, to)

	if src, err = os.Open(from); err != nil {
		slog.Colorized("      {red}fail{reset}  " + to)
		return err
	}
	defer src.Close()

	if dst, err = os.OpenFile(target, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666); err != nil {
		slog.Colorized("    {red}exists{reset}  " + to)
		return err
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		slog.Colorized("      {red}fail{reset}  " + to)
		return err
	}

	slog.Colorized("    {brightgreen}create{reset}  " + to)
	return nil
}

func zeusInit() {
	binaryPath := os.Args[0]
	gemDir := path.Dir(path.Dir(binaryPath))
	jsonPath := path.Join(gemDir, "examples/custom_plan/zeus.json")
	planPath := path.Join(gemDir, "examples/custom_plan/custom_plan.rb")
	copyFile(jsonPath, "zeus.json")
	copyFile(planPath, "custom_plan.rb")
}

func zeusCommands() {
	tree := config.BuildProcessTree()
	for _, command := range tree.Commands {
		alia := strings.Join(command.Aliases, ", ")
		var aliasPart string
		if len(alia) > 0 {
			aliasPart = " (alias: " + alia + ")"
		}
		println("zeus " + command.Name + aliasPart)
	}
}

func commandNotFound(command string) {
	println(red() + "Could not find command \"" + command + "\"." + reset())
}

func commandSpecificHelp(args []string) {
	if args[1] == "start" {
		execManPage("zeus-start")
	} else if args[1] == "init" {
		execManPage("zeus-init")
	} else {
		println(red() + "Command-level help is not yet fully implemented." + reset())
	}
}

func generalHelpRequested(args []string) bool {
	helps := []string{"help", "--help", "-h", "--help", "-?", "?"}
	if len(args) == 1 {
		for _, str := range helps {
			if args[0] == str {
				return true
			}
		}
	}
	return false
}

func printVersion() {
	println("Zeus version " + zeusversion.VERSION)
}
