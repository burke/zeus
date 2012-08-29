package shinylog

import "fmt"

var suppress bool = false

func Suppress() {
	suppress = true
}

func Red(msg string) {
	if !suppress {
		fmt.Println("\x1b[31m" + msg + "\x1b[0m")
	}
}

func Green(msg string) {
	if !suppress {
		fmt.Println("\x1b[32m" + msg + "\x1b[0m")
	}
}

func Yellow(msg string) {
	if !suppress {
		fmt.Println("\x1b[33m" + msg + "\x1b[0m")
	}
}

func Blue(msg string) {
	if !suppress {
		fmt.Println("\x1b[34m" + msg + "\x1b[0m")
	}
}

func Magenta(msg string) {
	if !suppress {
		fmt.Println("\x1b[35m" + msg + "\x1b[0m")
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
