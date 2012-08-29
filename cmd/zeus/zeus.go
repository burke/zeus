package main

import (
	"os"
	"syscall"

	"github.com/burke/zeus/zeusmaster"
	"github.com/burke/zeus/zeusclient"
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
	} else if args[0] == "start" {
		zeusmaster.Run(color)
	} else if args[0] == "init" {
		zeusInit()
	} else if args[0] == "commands" {
		zeusCommands()
	} else {
		tree := zeusmaster.BuildProcessTree()
		for name, _ := range tree.CommandsByName {
			if args[0] == name {
				zeusclient.Run(color)
				return
			}
		}

		commandNotFound(args[0])
	}
}

func execManPage(page string) {
	path, _:= os.Getwd()
	zeus := string(path) + "/man-comp/" + page
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
	println(red() + "zeus-init is not yet implemented." + reset())
}

func zeusCommands() {
	tree := zeusmaster.BuildProcessTree()
	for name, _ := range tree.CommandsByName {
		println("zeus " + name)
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
