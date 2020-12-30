package config

import (
	"errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type SystemConfig struct {
	UpdaterServer	string
	UpdaterPort		uint
	LocalName		string
	LaunchPath		string
	LaunchArg		string
	DownloadPath	string
}

type YamlEntry map[string]interface{}
type YamlConfig struct {
	Server YamlEntry `yaml:"server"`
	Local YamlEntry `yaml:"local"`
}

var (
	systemConfig SystemConfig
)

func GetSystemConfig() *SystemConfig {
	return &systemConfig
}

func (args *SystemConfig) DefaultConfig() {
	args.LaunchPath 	= "FilsoInstaller.exe"
	args.LaunchArg		= ""
	args.DownloadPath	= "__update"
	args.UpdaterServer 	= "localhost"
	args.UpdaterPort 	= 2008
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
	address := data.Server["address"]
	if address != nil {
		args.UpdaterServer = address.(string)
	} else {
		return errors.New("No address field")
	}

	port := data.Server["port"]
	if port != nil {
		args.UpdaterPort = uint(port.(int))
	} else {
		return errors.New("No port field")
	}

	name := data.Local["name"]
	if name != nil {
		args.LocalName = name.(string)
	} else {
		return errors.New("No Local name field")
	}

	launchPath := data.Local["launch"]
	if launchPath != nil {
		args.LaunchPath = launchPath.(string)
	} else {
		return errors.New("No launch file path field")
	}

	launchArg := data.Local["args"]
	if launchArg != nil {
		args.LaunchArg = launchArg.(string)
	} else {
		args.LaunchArg = ""
	}

	downloadPath := data.Local["download"]
	if downloadPath != nil {
		args.DownloadPath = downloadPath.(string)
	} else {
		args.DownloadPath = "__update"
	}

	return nil
}