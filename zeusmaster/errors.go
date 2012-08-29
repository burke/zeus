package zeusmaster

import (
	"fmt"
	slog "github.com/burke/zeus/shinylog"
)

var suppressErrors bool = false

func SuppressErrors() {
	suppressErrors = true
}

var (
	red = "\x1b[31m"
	yellow = "\x1b[33m"
	reset = "\x1b[0m"
)

func DisableErrorColor() {
	red = ""
	yellow = ""
	reset = ""
}

func ErrorConfigCommandCouldntStart(output string) {
	if !suppressErrors {
		slog.Red("Failed to initialize application from " + yellow + "zeus.json" + red + ".")
		slog.Red("The json file is valid, but the " + yellow + "command" + red + " could not be started:")
		fmt.Println(output)
		ExitNow(1)
	}
}

func ErrorConfigCommandCrashed(output string) {
	if !suppressErrors {
		slog.Red("Failed to initialize application from " + yellow + "zeus.json" + red + ".")
		slog.Red("The json file is valid, but the " + yellow + "command" + red + " terminated with this output:")
		fmt.Println(output)
		ExitNow(1)
	}
}

func ErrorConfigFileInvalidJson() {
	if !suppressErrors {
		slog.Red("The config file " + yellow + "zeus.json" + red + " contains invalid JSON and could not be parsed.")
		ExitNow(1)
	}
}

func ErrorConfigFileInvalidFormat() {
	if !suppressErrors {
		slog.Red("The config file " + yellow + "zeus.json" + red + " is not in the correct format.")
		ExitNow(1)
	}
}

func ErrorCantCreateListener() {
	if !suppressErrors {
		slog.Red("It looks like Zeus is already running. If not, remove " + yellow + ".zeus.sock" + red + " and try again.")
		ExitNow(1)
	}
}

func errorUnableToAcceptSocketConnection() {
	if !suppressErrors {
		slog.Red("Unable to accept socket connection.")
	}
}
