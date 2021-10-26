package main

import (
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

type LocalConfig struct {
	APIKey  string
	DocRoot string
}

func ReadLocalConfig(path string) (*LocalConfig, error) {
	cfg := &LocalConfig{}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	} else {
		err = yaml.Unmarshal(data, cfg)
		if err != nil {
			return nil, err
		}
	}
	env := os.Getenv("API_KEY")
	if env != "" {
		cfg.APIKey = env
	}
	env = os.Getenv("DOC_ROOT")
	if env != "" {
		cfg.DocRoot = env
	}
	return cfg, nil
}
