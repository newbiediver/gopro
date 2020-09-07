package config

import (
	"errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type SystemConfig struct {
	Port			uint
	ContentsPath	string
}

type YamlEntry map[string]interface{}
type YamlConfig struct {
	System YamlEntry `yaml:"system"`
}

var (
	systemConfig SystemConfig
)

func GetSystemConfig() *SystemConfig {
	return &systemConfig
}

func (args *SystemConfig) DefaultConfig() {
	args.ContentsPath 	= "./contents"
	args.Port 			= 2008
}

func (args *SystemConfig) LoadYaml(filename string) (*YamlConfig, bool) {
	config := new(YamlConfig)
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, false
	}

	err = yaml.Unmarshal(bytes, config)
	if err != nil {
		return nil, false
	}

	return config, true
}

func (args *SystemConfig) SetConfig(data *YamlConfig) error {
	port := data.System["port"]
	if port != nil {
		args.Port = uint(port.(int))
	} else {
		return errors.New("No port field")
	}

	contentsPath := data.System["contents"]
	if contentsPath != nil {
		args.ContentsPath = contentsPath.(string)
	} else {
		return errors.New("No contents path field")
	}

	return nil
}