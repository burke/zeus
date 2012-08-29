package zeusmaster

import (
	"encoding/json"
	"os"
	"bufio"
	"io/ioutil"
)

const configFile string = "zeus.json"

type config struct {
	Command string
	Plan interface{}
	Items map[string]string
}

func BuildProcessTree() (*ProcessTree) {
	conf := parseConfig()
	tree := &ProcessTree{}
	tree.CommandsByName = make(map[string]*CommandNode)
	tree.SlavesByName   = make(map[string]*SlaveNode)

	tree.ExecCommand = conf.Command

	plan, ok := conf.Plan.(map[string]interface{})
	if !ok {
		ErrorConfigFileInvalidFormat()
	}
	iteratePlan(tree, plan, nil)

	return tree
}

func iteratePlan(tree *ProcessTree, plan map[string]interface{}, parent *SlaveNode) {
	for name, v := range plan {
		if subPlan, ok := v.(map[string]interface{}); ok {
			newNode := tree.NewSlaveNode(name, parent)
			if parent == nil {
				tree.Root = newNode
			} else {
				parent.Slaves = append(parent.Slaves, newNode)
			}
			iteratePlan(tree, subPlan, newNode)
		} else {
			var newNode *CommandNode
			if aliases, ok := v.([]interface{}); ok {
				strs := make([]string, len(aliases))
				for _, alias := range aliases {
					strs = append(strs, alias.(string))
				}
				newNode = tree.NewCommandNode(name, strs, parent)
			} else if v == nil {
				newNode = tree.NewCommandNode(name, nil, parent)
			} else {
				ErrorConfigFileInvalidFormat()
			}
			parent.Commands = append(parent.Commands, newNode)
		}
	}
}

func parseConfig() (c config) {
	var conf config

	contents, err := readFile(configFile)
	if err != nil {
		ErrorConfigFileInvalidJson()
		panic(err)
	}

	json.Unmarshal(contents, &conf)
	return conf
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
