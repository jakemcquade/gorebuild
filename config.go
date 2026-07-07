package main

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Port int `yaml:"port"`
	} `yaml:"server"`
	Projects map[string]Project `yaml:"projects"`
}

type Project struct {
	Path   string `yaml:"path"`
	Branch string `yaml:"branch"` // optional; if empty, any branch is accepted
}

var config Config

func loadConfig(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, &config)
}
