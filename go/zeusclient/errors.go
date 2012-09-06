package zeusclient

import (
	slog "github.com/burke/zeus/go/shinylog"
	"os"
)

var (
	red    = "\x1b[31m"
	yellow = "\x1b[33m"
	reset  = "\x1b[0m"
)

func DisableErrorColor() {
	red = ""
	yellow = ""
	reset = ""
}

func ErrorCantConnectToMaster() {
	slog.Red("Can't connect to master. Run " + yellow + "zeus start" + red + " first.")
	os.Exit(1)
}
