package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// From https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions
// Ref: https://medium.com/thedevproject/easy-de-serialisation-of-yaml-files-in-go-4557456b0a98
// Ref: https://github.com/rhysd/actionlint/blob/main/config.go
type Config struct {
	Jobs map[string]Job `yaml:"jobs,omitempty"`
}

type Job struct {
	Steps Step   `yaml:"steps,omitempty"`
	Name  string `yaml:"name,omitempty"`
}

type Step []struct {
	Name string
	Uses string `yaml:"uses"`
}

func parseConfig(b []byte, path string) (*Config, error) {
	var c Config
	if err := yaml.Unmarshal(b, &c); err != nil {
		msg := strings.ReplaceAll(err.Error(), "\n", " ")
		return nil, fmt.Errorf("could not parse config file %q: %s", path, msg)
	}
	return &c, nil
}

func readConfigFile(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read config file %q: %w", path, err)
	}
	return parseConfig(b, path)
}

func getActions(file string) []string {
	var steps []string
	log.Printf("[DEBUG] Finding keys in file: %v", file)
	c, err := readConfigFile(file)

	if err != nil {
		log.Printf("[ERROR] Unable to read file(%v): %v", file, err.Error())
	}
	for _, v := range c.Jobs {
		for _, step := range v.Steps {
			if len(step.Uses) > 0 {
				steps = append(steps, step.Uses)
			}
		}
	}
	return steps
}
