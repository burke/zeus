package shinylog

import (
	"fmt"
	"strings"
)

var suppress bool = false

var (
	red     = "\x1b[31m"
	green   = "\x1b[32m"
	yellow  = "\x1b[33m"
	blue    = "\x1b[34m"
	magenta = "\x1b[35m"
	reset   = "\x1b[0m"
)

func Suppress() {
	suppress = true
}

func DisableColor() {
	red = ""
	green = ""
	yellow = ""
	blue = ""
	magenta = ""
	reset = "\x1b[0m"
}

func Colorized(msg string) (printed bool) {
	if !suppress {
		msg = strings.Replace(msg, "{red}", red, -1)
		msg = strings.Replace(msg, "{green}", green, -1)
		msg = strings.Replace(msg, "{yellow}", yellow, -1)
		msg = strings.Replace(msg, "{blue}", blue, -1)
		msg = strings.Replace(msg, "{magenta}", magenta, -1)
		msg = strings.Replace(msg, "{reset}", reset, -1)
		fmt.Println(msg + reset)
	}
	return !suppress
}

func Red(msg string) bool {
	return Colorized("{red}" + msg)
}

func Green(msg string) bool {
	return Colorized("{green}" + msg)
}

func Yellow(msg string) bool {
	return Colorized("{yellow}" + msg)
}

func Blue(msg string) bool {
	return Colorized("{blue}" + msg)
}

func Magenta(msg string) bool {
	return Colorized("{magenta}" + msg)
}
