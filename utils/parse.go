package utils

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/exp/slog"

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

// ParseLookup parses the step string and returns the
// associated repository and ref.
func ParseLookup(s string) (repo string, ref string, ok bool) {
	org, repo, _, ref, err := parseActionName(s)
	if err == nil {
		url := fmt.Sprintf("https://github.com/%s/%s", org, repo)
		slog.Debug(fmt.Sprintf("parsed repo: %s, ref: %s", url, ref))
		return url, ref, true
	}

	slog.Warn(fmt.Sprintf("failed to parse action name: %s with err: %v", s, err))
	if !strings.HasPrefix(s, "https://github.com") {
		s, _ = url.JoinPath("https://github.com", s)
	}

	slog.Debug("parsed repo", s)
	if parts := strings.SplitN(s, "@", 2); len(parts) == 2 {
		return parts[0], parts[1], true
	}
	return s, "", true
}

func parseActionName(action string) (org, repo, path, ref string, err error) {
	r := regexp.MustCompile(`^([^/@]+)/([^/@]+)(/([^@]*))?(@(.*))?$`)
	matches := r.FindStringSubmatch(action)
	if len(matches) < 7 || matches[6] == "" {
		err = fmt.Errorf("invalid action name: %s", action)
		return
	}
	org = matches[1]
	repo = matches[2]
	path = matches[4]
	ref = matches[6]
	return
}
