package main

import (
	"os"
	"syscall"

	"github.com/burke/zeus/zeusmaster"
	"github.com/burke/zeus/zeusclient"
)

func main () {
	if generalHelpRequested() {
		showManPage("zeus")
	} else if os.Args[1] == "help" {
		commandSpecificHelp()
	} else if os.Args[1] == "start" {
		zeusmaster.Run()
	} else if os.Args[1] == "init" {
		zeusInit()
	} else if os.Args[1] == "commands" {
		zeusCommands()
	} else {
		tree := zeusmaster.BuildProcessTree()
		for name, _ := range tree.CommandsByName {
			if os.Args[1] == name {
				zeusclient.Run()
				return
			}
		}

		commandNotFound(os.Args[1])
	}
}

func showManPage(page string) {
	path, _:= os.Getwd()
	zeus := string(path) + "/man-comp/" + page
	syscall.Exec("/usr/bin/env", []string{"/usr/bin/env", "man", zeus}, os.Environ())
}

func zeusInit() {
	println("\x1b[31mzeus-init is not yet implemented.\x1b[0m")
}

func zeusCommands() {
	tree := zeusmaster.BuildProcessTree()
	for name, _ := range tree.CommandsByName {
		println("zeus " + name)
	}
}

func commandNotFound(command string) {
	println("\x1b[31mCould not find command \"" + command + "\".\x1b[0m")
}

func commandSpecificHelp() {
	if os.Args[2] == "start" {
		showManPage("zeus-start")
	} else if os.Args[2] == "init" {
		showManPage("zeus-init")
	} else {
		println("\x1b[31mCommand-level help is not yet fully implemented.\x1b[0m")
	}
}

func generalHelpRequested() bool {
	helps := []string{"help", "--help", "-h", "--help", "-?", "?"}
	if len(os.Args) == 1 {
		return true
	}

	for _, str := range helps {
		if os.Args[1] == str {
			return 2 == len(os.Args)
			return true
		}
	}
	return false
}
