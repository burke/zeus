package config

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/burke/zeus/go/processtree"
)

type config struct {
	Command string
	Plan    interface{}
	Items   map[string]string
}

func BuildProcessTree(configFile string, tree processtree.ProcessTree) error {
	plan, cmd, err := loadConfig(configFile)
	if err != nil {
		return err
	}

	for name, v := range plan {
		subPlan, ok := v.(map[string]interface{})
		if !ok {
			return fmt.Errorf("expected root node %s in plan to map to an object, got %v", name, v)
		}

		tree.AddRootNode(name, cmd)
		if err := iteratePlan(tree, subPlan, name); err != nil {
			return err
		}
	}

	return nil
}

func AllCommandsAndAliases(configFile string) (map[string][]string, error) {
	plan, _, err := loadConfig(configFile)
	if err != nil {
		return nil, err
	}

	cmds := make(map[string][]string)
	if err := planCommandsAndAliases(plan, cmds); err != nil {
		return nil, err
	}

	return cmds, nil
}

func planCommandsAndAliases(plan map[string]interface{}, cmds map[string][]string) error {
	for name, v := range plan {
		if subPlan, ok := v.(map[string]interface{}); ok {
			planCommandsAndAliases(subPlan, cmds)
		} else {
			if aliases, ok := v.([]interface{}); ok {
				strs := make([]string, len(aliases))
				for i, alias := range aliases {
					strs[i] = alias.(string)
				}
				cmds[name] = strs
			} else if v == nil {
				cmds[name] = []string{}
			} else {
				return fmt.Errorf("Expected command %q to have no value or be a list of aliases, got: %v", name, v)
			}
		}
	}

	return nil
}

func loadConfig(configFile string) (map[string]interface{}, []string, error) {
	conf, err := parseConfig(configFile)
	if err != nil {
		return nil, nil, err
	}

	plan, ok := conf.Plan.(map[string]interface{})
	if !ok {
		return nil, nil, errors.New("The config file must contain a `plan` key with an object value")
	}

	cmd := strings.Split(conf.Command, " ")

	return plan, cmd, nil
}

func iteratePlan(
	tree processtree.ProcessTree,
	plan map[string]interface{},
	parent string,
) error {
	for name, v := range plan {
		if subPlan, ok := v.(map[string]interface{}); ok {
			tree.AddChildNode(name, parent)
			iteratePlan(tree, subPlan, name)
		} else {
			if aliases, ok := v.([]interface{}); ok {
				strs := make([]string, len(aliases))
				for i, alias := range aliases {
					strs[i] = alias.(string)
				}
				tree.AddCommand(name, parent, strs)
			} else if v == nil {
				tree.AddCommand(name, parent, nil)
			} else {
				return fmt.Errorf("Expected command %q to have no value or be a list of aliases, got: %v", name, v)
			}
		}
	}

	return nil
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

func parseConfig(configFile string) (config, error) {
	var conf config

	contents, err := readConfigFileOrDefault(configFile)
	if err != nil {
		return conf, fmt.Errorf("The config file %s could not be read: %v", configFile, err)
	}
	if err := json.Unmarshal(contents, &conf); err != nil {
		return conf, fmt.Errorf("The config file %s could not be parsed: %v", configFile, err)
	}

	return conf, nil
}

func readFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(file)

	contents, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return contents, nil
}
