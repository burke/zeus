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
	"github.com/burke/zeus/go/filemonitor"
	slog "github.com/burke/zeus/go/shinylog"
	"github.com/burke/zeus/go/zeusclient"
	"github.com/burke/zeus/go/zeusmaster"
	"github.com/burke/zeus/go/zeusversion"
)

var color bool = true

func main() {
	args := os.Args[1:]
	configFile := "zeus.json"
	fileChangeDelay := filemonitor.DefaultFileChangeDelay

	for ; args != nil && len(args) > 0 && args[0][0] == '-'; args = args[1:] {
		switch args[0] {
		case "--no-color":
			color = false
			slog.DisableColor()
		case "--log":
			tracefile, err := os.OpenFile(args[1], os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
			if err == nil {
				slog.SetTraceLogger(slog.NewTraceLogger(tracefile))
				args = args[1:]
			} else {
				fmt.Printf("Could not open trace file %s\n", args[1])
				return
			}
		case "--file-change-delay":
			if len(args) > 1 {
				delay, err := time.ParseDuration(args[1])
				if err != nil {
					execManPage("zeus")
				}
				args = args[1:]
				fileChangeDelay = delay
			} else {
				execManPage("zeus")
			}
		case "--config":
			_, err := os.Stat(args[1])
			if err != nil {
				fmt.Printf("Config file doesn't exist: %s (%e)\n", args[1], err)
				return
			}
			configFile = args[1]
			args = args[1:]
		case "--version":
			printVersion()
			return
		}
	}
	if len(args) == 0 {
		execManPage("zeus")
		return
	}

	if generalHelpRequested(args) {
		execManPage("zeus")
	} else if args[0] == "help" {
		commandSpecificHelp(args)
	} else if args[0] == "version" {
		printVersion()
	} else if args[0] == "start" {
		os.Exit(zeusmaster.Run(configFile, fileChangeDelay))
	} else if args[0] == "init" {
		zeusInit()
	} else if args[0] == "commands" {
		zeusCommands(configFile)
	} else {
		exists, err := commandExists(configFile, args[0])
		if err != nil {
			fmt.Printf("%s%s%s", red(), err, reset())
			os.Exit(1)
		}
		if !exists {
			commandNotFound(args[0])
			os.Exit(1)
		}

		// Don't confuse the master by sending *full* args to
		// it; just those that are not zeus-specific.
		os.Exit(zeusclient.Run(args, os.Stdin, os.Stdout))
	}
}

func commandExists(configFile, name string) (bool, error) {
	cmds, err := config.AllCommandsAndAliases(configFile)
	if err != nil {
		return false, err
	}

	if _, ok := cmds[name]; ok {
		return true, nil
	}

	for _, aliases := range cmds {
		for _, alias := range aliases {
			if name == alias {
				return true, nil
			}
		}
	}

	return false, nil
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

func zeusCommands(configFile string) {
	cmds, err := config.AllCommandsAndAliases(configFile)
	if err != nil {
		fmt.Printf("%s%s%s", red(), err, reset())
		os.Exit(1)
	}

	for name, aliases := range cmds {
		alia := strings.Join(aliases, ", ")
		var aliasPart string
		if len(alia) > 0 {
			aliasPart = " (alias: " + alia + ")"
		}
		fmt.Println("zeus " + name + aliasPart)
	}
}

func commandNotFound(command string) {
	fmt.Println(red() + "Could not find command \"" + command + "\"." + reset())
}

func commandSpecificHelp(args []string) {
	if args[1] == "start" {
		execManPage("zeus-start")
	} else if args[1] == "init" {
		execManPage("zeus-init")
	} else {
		fmt.Println(red() + "Command-level help is not yet fully implemented." + reset())
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
	fmt.Println("Zeus version " + zeusversion.VERSION)
}
