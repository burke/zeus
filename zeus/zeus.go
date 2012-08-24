package zeus

import (
	"time"

	ptree "github.com/burke/zeus/processtree"
	"github.com/burke/zeus/config"
	"github.com/burke/zeus/slavemonitor"
	"github.com/burke/zeus/clienthandler"
	"github.com/burke/zeus/filemonitor"
)


func Pseudo() {
	// build the tree before anything else happens
	var tree *ptree.ProcessTree
	tree = config.BuildTree()
	go slavemonitor.Run(tree)
	go clienthandler.Run(tree)
	go filemonitor.Run(tree)

	time.Sleep(500 * time.Millisecond)
}


