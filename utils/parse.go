package utils

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type GHActionSpec struct {
	Outputs map[string]interface{} `yaml:"outputs,omitempty"`
}

// ParseActionOutputs locates `action.yml` or `action.yaml` in `root` and returns all top-level outputs.
func ParseActionOutputs(root string) ([]string, error) {
	ymlPath := filepath.Join(root, "action.yml")
	yamlPath := filepath.Join(root, "action.yaml")

	var actionFile string
	switch {
	case fileExists(ymlPath):
		actionFile = ymlPath
	case fileExists(yamlPath):
		actionFile = yamlPath
	default:
		logrus.Warnf("action.yml or action.yaml not found in %s. Skipping output variable processing.", root)
		return []string{}, nil
	}

	raw, err := os.ReadFile(actionFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read action file: %w", err)
	}

	var spec GHActionSpec
	if err := yaml.Unmarshal(raw, &spec); err != nil {
		return nil, fmt.Errorf("failed to parse action.yml: %w", err)
	}

	keys := make([]string, 0, len(spec.Outputs))
	for k := range spec.Outputs {
		keys = append(keys, k)
	}
	return keys, nil
}

// fileExists checks if the given path exists and is a file.
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}
