package main

import (
	"os"
	"strings"
	"syscall"
	"path"
	"io"

	"github.com/burke/zeus/go/zeusmaster"
	"github.com/burke/zeus/go/zeusclient"
	"github.com/burke/zeus/go/zeusversion"
)

var color bool = true

func main () {
	if len(os.Args) == 1 {
		execManPage("zeus")
	}

	var args []string
	if os.Args[1] == "--no-color" {
		color = false
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
		zeusmaster.Run(color)
	} else if args[0] == "init" {
		zeusInit()
	} else if args[0] == "commands" {
		zeusCommands()
	} else {
		tree := zeusmaster.BuildProcessTree()
		for _, name := range tree.AllCommandsAndAliases() {
			if args[0] == name {
				zeusclient.Run(color)
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

func zeusInit() {
	binaryPath := os.Args[0]
	gemDir := path.Dir(path.Dir(binaryPath))
	jsonPath := path.Join(gemDir, "examples/zeus.json")
	wd, _ := os.Getwd()
	targetPath := path.Join(wd, "zeus.json")
	src, err := os.Open(jsonPath)
	if err != nil {
		println(red() + "Could not open template file" + reset())
		return
	}
	defer src.Close()

	dst, err := os.OpenFile(targetPath, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
	if err != nil {
		println(red() + "Could not create zeus.json" + reset())
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	if err != nil {
		println(red() + "Could not write to zeus.json" + reset())
	}
	println("Wrote zeus.json")
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
