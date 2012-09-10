package zeusmaster

import (
	"fmt"
	slog "github.com/burke/zeus/go/shinylog"
	"os"
)

func ErrorConfigCommandCouldntStart(output string) {
	slog.Red("Failed to initialize application from {yellow}zeus.json{red}.")
	if slog.Red("The json file is valid, but the {yellow}command{red} could not be started:") {
		fmt.Println(output)
		ExitNow(1)
	}
}

func ErrorConfigCommandCrashed(output string) {
	slog.Red("Failed to initialize application from {yellow}zeus.json{red}.")
	if slog.Red("The json file is valid, but the {yellow}command{red} terminated with this output:") {
		fmt.Println(output)
		ExitNow(1)
	}
}

// The config file is loaded before any goroutines are launched that require cleanup,
// and our exitNow goroutine has not been spawned yet, so we will just explicitly exit
// in the json-related errors..
func ErrorConfigFileMissing() {
	if slog.Red("Required config file {yellow}zeus.json{red} not found in the current directory.") {
		os.Exit(1)
	}
}

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
	if slog.Red("It looks like Zeus is already running. If not, remove {yellow}.zeus.sock{red} and try again.") {
		ExitNow(1)
	}
}

func errorUnableToAcceptSocketConnection() {
	slog.Red("Unable to accept socket connection.")
}

func errorFailedReadFromWatcher(err error) {
	slog.Red("Failed to read from FileSystem watcher process: " + err.Error())
}

func ErrorFileMonitorWrapperCrashed(err error) {
	if slog.Red("The FileSystem watcher process crashed: " + err.Error()) {
		ExitNow(1)
	}
}
