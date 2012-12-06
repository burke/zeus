package main

import (
	"io"
	"os"
	"path"
	"strings"
	"syscall"

	slog "github.com/burke/zeus/go/shinylog"
	"github.com/burke/zeus/go/zeusclient"
	"github.com/burke/zeus/go/zeusmaster"
	"github.com/burke/zeus/go/zeusversion"
)

var color bool = true

func main() {
	if len(os.Args) == 1 {
		execManPage("zeus")
	}

	var args []string
	if os.Args[1] == "--no-color" {
		color = false
		slog.DisableColor()
		args = os.Args[2:]
	} else {
		args = os.Args[1:]
	}

	if generalHelpRequested(args) {
		execManPage("zeus")
	} else if args[0] == "help" {
		commandSpecificHelp(args)
	} else if args[0] == "version" || args[0] == "--version" {
		println("Zeus version " + zeusversion.VERSION)
	} else if args[0] == "start" {
		zeusmaster.Run()
	} else if args[0] == "init" {
		zeusInit()
	} else if args[0] == "commands" {
		zeusCommands()
	} else {
		tree := zeusmaster.BuildProcessTree()
		for _, name := range tree.AllCommandsAndAliases() {
			if args[0] == name {
				zeusclient.Run()
				return
			}
		}

		commandNotFound(args[0])
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
	tree := zeusmaster.BuildProcessTree()
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
