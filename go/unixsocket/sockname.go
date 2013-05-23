package unixsocket

import "os"

var sockName string

func init() {
	sockName = os.Getenv("ZEUSSOCK")
	if sockName == "" {
		sockName = ".zeus.sock"
	}
}

func ZeusSockName() string {
	return sockName
}
