package shinylog

import "fmt"

var suppress bool = false
var color = true

const (
	red = "\x1b[31m"
	green = "\x1b[32m"
	yellow = "\x1b[33m"
	blue = "\x1b[34m"
	magenta = "\x1b[35m"
	reset = "\x1b[0m"
)

func Suppress() {
	suppress = true
}

func DisableColor() {
	color = false
}

func Red(msg string) {
	if !suppress {
		if color {
			fmt.Println(red + msg + reset)
		} else {
			fmt.Println(msg)
		}
	}
}

func Green(msg string) {
	if !suppress {
		if color {
			fmt.Println(green + msg + reset)
		} else {
			fmt.Println(msg)
		}
	}
}

func Yellow(msg string) {
	if !suppress {
		if color {
			fmt.Println(yellow + msg + reset)
		} else {
			fmt.Println(msg)
		}
	}
}

func Blue(msg string) {
	if !suppress {
		if color {
			fmt.Println(blue + msg + reset)
		} else {
			fmt.Println(msg)
		}
	}
}

func Magenta(msg string) {
	if !suppress {
		if color {
			fmt.Println(magenta + msg + reset)
		} else {
			fmt.Println(msg)
		}
	}
}


func SlaveBooted(name string) {
	Blue("ready   : " + name)
}

func SlaveKilled(name string) {
	Yellow("killed  : " + name)
}

func SlaveDied(name string) {
	Yellow("died    : " + name)
}

func StartingZeus() {
	if color {
		fmt.Println("\x1b[32mStarting \x1b[33mZ\x1b[31me\x1b[34mu\x1b[35ms\x1b[32m server\x1b[0m")
	} else {
		fmt.Println("Starting Zeus server")
	}
}
