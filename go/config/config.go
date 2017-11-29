package config

import (
	"bufio"
	"encoding/json"
	"io/ioutil"
	"os"
	"path"

	"github.com/burke/zeus/go/filemonitor"
	"github.com/burke/zeus/go/processtree"
	"github.com/burke/zeus/go/zerror"
)

type config struct {
	Command string
	Plan    interface{}
	Items   map[string]string
}

func BuildProcessTree(configFile string, monitor filemonitor.FileMonitor) *processtree.ProcessTree {
	conf := parseConfig(configFile)
	tree := &processtree.ProcessTree{}
	tree.WorkersByName = make(map[string]*processtree.WorkerNode)
	tree.StateChanged = make(chan bool, 16)

	tree.ExecCommand = conf.Command

	plan, ok := conf.Plan.(map[string]interface{})
	if !ok {
		zerror.ErrorConfigFileInvalidFormat()
	}
	iteratePlan(tree, plan, monitor, nil)

	return tree
}

func iteratePlan(
	tree *processtree.ProcessTree,
	plan map[string]interface{},
	monitor filemonitor.FileMonitor,
	parent *processtree.WorkerNode,
) {
	for name, v := range plan {
		if subPlan, ok := v.(map[string]interface{}); ok {
			newNode := tree.NewWorkerNode(name, parent, monitor)
			if parent == nil {
				tree.Root = newNode
			} else {
				parent.Workers = append(parent.Workers, newNode)
			}
			iteratePlan(tree, subPlan, monitor, newNode)
		} else {
			var newNode *processtree.CommandNode
			if aliases, ok := v.([]interface{}); ok {
				strs := make([]string, len(aliases))
				for i, alias := range aliases {
					strs[i] = alias.(string)
				}
				newNode = tree.NewCommandNode(name, strs, parent)
			} else if v == nil {
				newNode = tree.NewCommandNode(name, nil, parent)
			} else {
				zerror.ErrorConfigFileInvalidFormat()
			}
			parent.Commands = append(parent.Commands, newNode)
		}
	}
}

func defaultConfigPath() string {
	binaryPath := os.Args[0]
	gemDir := path.Dir(path.Dir(binaryPath))
	jsonpath := path.Join(gemDir, "examples/zeus.json")
	return jsonpath
}

func readConfigFileOrDefault(configFile string) ([]byte, error) {
	contents, err := readFile(configFile)
	if err != nil {
		switch err.(type) {
		case *os.PathError:
			return readFile(defaultConfigPath())
		default:
			return contents, err
		}
	}
	return contents, err
}

func parseConfig(configFile string) (c config) {
	var conf config

	contents, err := readConfigFileOrDefault(configFile)
	if err != nil {
		zerror.ErrorConfigFileInvalidJson()
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
