package zeusmaster

import (
	"goyaml"
)

type config struct {
	Command string
	Plan interface{}
	Items map[string]string
}

var zfile = "---\ncommand: \"/Users/burke/.rbenv/shims/ruby /Users/burke/go/src/github.com/burke/zeus/a.rb\"\nplan:\n  default_bundle:\n    development_environment:\n      prerake:\n        rake:\n      runner:\n      console:\n    test_environment:\n      testtask:\n      test_helper:\n        testrb:\n      spec_helper:\n        rspec:\n\nitems:\n  default_bundle: |\n    require 'rails/all'\n  development_environment: |\n    Bundler.require(:development)\n  test_environment: |\n    Bundler.require(:test)\n  runner: |\n    omg\n\n\n"

func parseConfig() (c config) {
	var conf config
	goyaml.Unmarshal([]byte(zfile), &conf)
	return conf
}

func BuildProcessTree() (*ProcessTree) {
	conf := parseConfig()
	tree := &ProcessTree{}

	tree.ExecCommand = conf.Command

	return tree
}
