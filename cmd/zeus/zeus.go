package main

import (
	"os"

	"github.com/burke/zeus/zeusmaster"
	"github.com/burke/zeus/zeusclient"
)

func main () {
	if len(os.Args) > 1 && os.Args[1] == "start" {
		zeusmaster.Run()
	} else {
		zeusclient.Run()
	}
}
