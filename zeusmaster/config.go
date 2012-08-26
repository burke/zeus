package zeusmaster

import (
	"goyaml"
	"os"
	"bufio"
	"io/ioutil"
)

type config struct {
	Command string
	Plan interface{}
	Items map[string]string
}

func parseConfig() (c config) {
	var conf config

	contents, err := readFile("zeus.yml")
	if err != nil {
		panic(err)
	}

	goyaml.Unmarshal(contents, &conf)
	return conf
}

func BuildProcessTree() (*ProcessTree) {
	conf := parseConfig()
	tree := &ProcessTree{}

	tree.ExecCommand = conf.Command

	return tree
}

func readFile(path string) (contents []byte, err error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(file)

	contents, err = ioutil.ReadAll(reader)
	return
}
