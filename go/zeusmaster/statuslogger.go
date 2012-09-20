package zeusmaster

import (
	slog "github.com/burke/zeus/go/shinylog"
)

func StatusUpdate(identifier, state string) {
	switch state {
	case sWaiting:
		slog.Yellow("[waiting]    " + identifier)
	case sUnbooted:
	case sBooting:
		slog.Blue("[running]    " + identifier)
	case sCrashed:
		slog.Red("[crashed]    " + identifier)
	case sReady:
		slog.Green("[ready]      " + identifier)
	}
}
