package zeusmaster

import (
	"fmt"
	slog "github.com/burke/zeus/shinylog"
)

var suppressErrors bool = false

func SuppressErrors() {
	suppressErrors = true
}

func ErrorConfigCommandCouldntStart(output string) {
	if !suppressErrors {
		slog.Red("Failed to initialize application from \x1b[33mzeus.json\x1b[31m.")
		slog.Red("The json file is valid, but the \x1b[33mcommand\x1b[31m could not be started:")
		fmt.Println(output)
		ExitNow(1)
	}
}

func ErrorConfigCommandCrashed(output string) {
	if !suppressErrors {
		slog.Red("Failed to initialize application from \x1b[33mzeus.json\x1b[31m.")
		slog.Red("The json file is valid, but the \x1b[33mcommand\x1b[31m terminated with this output:")
		fmt.Println(output)
		ExitNow(1)
	}
}

func ErrorConfigFileInvalidJson() {
	if !suppressErrors {
		slog.Red("The config file \x1b[33mzeus.json\x1b[31m contains invalid JSON and could not be parsed.")
		ExitNow(1)
	}
}

func ErrorConfigFileInvalidFormat() {
	if !suppressErrors {
		slog.Red("The config file \x1b[33mzeus.json\x1b[31m is not in the correct format.")
		ExitNow(1)
	}
}

func ErrorCantCreateListener() {
	if !suppressErrors {
		slog.Red("It looks like Zeus is already running. If not, remove \x1b[33m.zeus.sock\x1b[31m and try again.")
		ExitNow(1)
	}
}

func errorUnableToAcceptSocketConnection() {
	if !suppressErrors {
		slog.Red("Unable to accept socket connection.")
	}
}
