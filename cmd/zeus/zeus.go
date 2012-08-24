package main

import (
	"github.com/burke/zeus/zeus"
)

func main () {

	conf := zeus.ParseConfig()

	println(conf.Command)

	zeus.Run(conf)
}
