package zeusmaster

import (
	"fmt"
	"os"

	slog "github.com/burke/zeus/go/shinylog"
)

func Error(msg string) {
	ExitNow(1, func() {
		slog.Red(msg)
	})
}

func ErrorConfigCommandCouldntStart(msg, output string) {
	ExitNow(1, func() {
		slog.Red("Failed to initialize application from {yellow}zeus.json{red}.")
		slog.Red("The json file is valid, but the {yellow}command{red} could not be started:\n\x1b[0m" + output)
	})
}

func ErrorConfigCommandCrashed(output string) {
	ExitNow(1, func() {
		slog.Red("Couldn't boot application. {yellow}command{red} terminated with this output:")
		fmt.Println(output)
	})
}

// The config file is loaded before any goroutines are launched that require cleanup,
// and our exitNow goroutine has not been spawned yet, so we will just explicitly exit
// in the json-related errors..
func ErrorConfigFileInvalidJson() {
	if slog.Red("The config file {yellow}zeus.json{red} contains invalid JSON and could not be parsed.") {
		os.Exit(1)
	}
}

func ErrorConfigFileInvalidFormat() {
	if slog.Red("The config file {yellow}zeus.json{red} is not in the correct format.") {
		os.Exit(1)
	}
}

func ErrorCantCreateListener() {
	ExitNow(1, func() {
		slog.Red("It looks like Zeus is already running. If not, remove {yellow}.zeus.sock{red} and try again.")
	})
}

func errorUnableToAcceptSocketConnection() {
	slog.Red("Unable to accept socket connection.")
}
