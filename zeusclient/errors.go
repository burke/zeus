package zeusclient

import (
	"os"
	slog "github.com/burke/zeus/shinylog"
)

func ErrorCantConnectToMaster() {
	slog.Red("Can't connect to master. Run \x1b[33mzeus start\x1b[31m first.")
	os.Exit(1)
}
