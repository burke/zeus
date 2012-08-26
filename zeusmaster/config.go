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

func BuildProcessTree() (*ProcessTree) {
	conf := parseConfig()
	tree := &ProcessTree{}
	tree.nodesByName = make(map[string]*ProcessTreeNode)

	tree.ExecCommand = conf.Command

	plan, ok := conf.Plan.(map[interface{}]interface{})
	if !ok {
		panic("invalid config file")
	}

	iteratePlan(tree, plan, nil)
	iterateItems(tree, conf.Items)

	return tree
}

func iterateItems(tree *ProcessTree, items map[string]string) {
	for name, action := range items {
		node := tree.FindNodeByName(name)
		if node == nil {
			panic("No map entry for " + name)
		}
		node.Action = action
	}
}

// so much ugly casting to deal with generic yaml. So much pain.
// surely there's a nicer way to do this.
func iteratePlan(tree *ProcessTree, plan map[interface{}]interface{}, parent *SlaveNode) {
	for k, v := range plan {
		name, ok := k.(string)
		if !ok {
			panic("key not a string")
		}

		if subPlan, ok := v.(map[interface{}]interface{}); ok {
			newNode := tree.NewSlaveNode(name)
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
				newNode = tree.NewCommandNode(name, strs)
			} else {
				if v == nil {
					newNode = tree.NewCommandNode(name, nil)
				} else {
					panic("Invalid config file")
				}
			}
			parent.Commands = append(parent.Commands, newNode)
		}
	}
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


func readFile(path string) (contents []byte, err error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(file)

	contents, err = ioutil.ReadAll(reader)
	return
}
