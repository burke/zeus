package shinylog

import "fmt"

func Red(msg string) {
	fmt.Println("\x1b[31m" + msg + "\x1b[0m")
}

func Green(msg string) {
	fmt.Println("\x1b[32m" + msg + "\x1b[0m")
}

func Yellow(msg string) {
	fmt.Println("\x1b[33m" + msg + "\x1b[0m")
}

func Blue(msg string) {
	fmt.Println("\x1b[34m" + msg + "\x1b[0m")
}

func Magenta(msg string) {
	fmt.Println("\x1b[35m" + msg + "\x1b[0m")
}
