package zeusclient

import (
	slog "github.com/burke/zeus/go/shinylog"
)

func errorCantConnectToMaster() {
	slog.Red("Can't connect to master. Run {yellow}zeus start{red} first.\r")
}
