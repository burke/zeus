package main

import (
	"testing"

	"github.com/burke/zeus/go/config"
)

func TestDefaultConfigFile(t *testing.T) {
  if config.ConfigFile != "zeus.json" {
    t.Fail()
  }
}

