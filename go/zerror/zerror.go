package zerror

import (
	"fmt"
	"os"
	"syscall"

	slog "github.com/burke/zeus/go/shinylog"
)

var finalOutput []func()

func Init() {
	finalOutput = make([]func(), 0)
}

func PrintFinalOutput() {
	for _, cb := range finalOutput {
		cb()
	}
}

// TODO: this is gross because code is ignored.
func ExitNow(code int, finalOuputCallback func()) {
	finalOutput = append(finalOutput, finalOuputCallback)
	proc, _ := os.FindProcess(os.Getpid())
	proc.Signal(syscall.SIGTERM)
}

func Error(msg string) {
	ExitNow(1, func() {
		slog.Red(msg)
	})
}

func ErrorCantConnectToMaster() {
	slog.StdErrorString("Can't connect to master. Run {yellow}zeus start{red} first.\r")
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
func ErrorConfigFileInvalidJSON() {
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

func ErrorUnableToAcceptSocketConnection() {
	slog.Red("Unable to accept socket connection.")
}
