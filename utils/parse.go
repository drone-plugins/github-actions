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

// ActionDir joins cloneDir with an optional action subdirectory.
// Returns an error if actionPath escapes cloneDir.
func ActionDir(cloneDir, actionPath string) (string, error) {
	if actionPath == "" {
		return cloneDir, nil
	}

	clean := filepath.Clean(actionPath)
	if clean == "." || clean == "" {
		return cloneDir, nil
	}
	if filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("invalid action path: %s", actionPath)
	}

	actionDir := filepath.Join(cloneDir, clean)
	rel, err := filepath.Rel(cloneDir, actionDir)
	if err != nil {
		return "", fmt.Errorf("invalid action path: %w", err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("action path escapes clone dir: %s", actionPath)
	}
	return actionDir, nil
}

// ParseLookup parses the step string and returns the
// associated repository, ref, and optional action subdirectory path.
func ParseLookup(s string) (repo string, ref string, path string, ok bool) {
	org, repoName, actionPath, ref, err := parseActionName(s)
	if err == nil {
		url := fmt.Sprintf("https://github.com/%s/%s", org, repoName)
		slog.Debug(fmt.Sprintf("parsed repo: %s, ref: %s, path: %s", url, ref, actionPath))
		return url, ref, actionPath, true
	}

	slog.Warn(fmt.Sprintf("failed to parse action name: %s with err: %v", s, err))
	if !strings.HasPrefix(s, "https://github.com") {
		s, _ = url.JoinPath("https://github.com", s)
	}

	slog.Debug("parsed repo", s)
	if parts := strings.SplitN(s, "@", 2); len(parts) == 2 {
		return parts[0], parts[1], "", true
	}
	return s, "", "", true
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
