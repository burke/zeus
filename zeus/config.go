package zeus

import (
	"goyaml"
)

type Config struct {
	Command string
	Plan interface{}
	Items map[string]string
}

var zfile = "---\ncommand: \"/Users/burke/.rbenv/shims/ruby /Users/burke/go/src/github.com/burke/zeus/a.rb\"\nplan:\n  default_bundle:\n    development_environment:\n      prerake:\n        rake:\n      runner:\n      console:\n    test_environment:\n      testtask:\n      test_helper:\n        testrb:\n      spec_helper:\n        rspec:\n\nitems:\n  default_bundle: |\n    require 'rails/all'\n  development_environment: |\n    Bundler.require(:development)\n  test_environment: |\n    Bundler.require(:test)\n  runner: |\n    omg\n\n\n"

func ParseConfig() (c Config) {
	var conf Config
	goyaml.Unmarshal([]byte(zfile), &conf)

	return conf
}

