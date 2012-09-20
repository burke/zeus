package zeusclient

import (
	slog "github.com/burke/zeus/go/shinylog"
	"os"
)

func ErrorCantConnectToMaster() {
	slog.Red("Can't connect to master. Run {yellow}zeus start{red} first.\r")
	os.Exit(1)
}
