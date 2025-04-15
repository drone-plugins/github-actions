package utils

import (
	"os"
	"runtime"
	"strings"

	"github.com/joho/godotenv"
	"github.com/pkg/errors"
)

func CreateEnvAndSecretFile(envFile, secretFile string, secrets []string) error {
	envVars := getEnvVars()

	actionEnvVars := make(map[string]string)
	for key, val := range envVars {
		if !strings.HasPrefix(key, "PLUGIN_") && !Exists(secrets, key) {
			actionEnvVars[key] = val
		}
	}

	arch := "X64"
	if runtime.GOARCH == "arm64" {
		arch = "arm64"
	}

	ostype := "Linux"
	runner_tool_cache := "/opt/hostedtoolcache"
	if runtime.GOOS == "darwin" {
		ostype = "macOS"
		runner_tool_cache = "/Users/anka/hostedtoolcache"
	} else if runtime.GOOS == "windows" {
		ostype = "Windows"
		runner_tool_cache = "C:\\hostedtoolcache\\windows"
	}

	// Map Drone variables to GitHub variables
	actionEnvVars = combineEnv(map[string]string{
		"GITHUB_BASE_REF":         envVars["DRONE_TARGET_BRANCH"],
		"GITHUB_HEAD_REF":         envVars["DRONE_SOURCE_BRANCH"],
		"GITHUB_REF":              envVars["DRONE_COMMIT_REF"],
		"GITHUB_REPOSITORY":       envVars["DRONE_REPO"],
		"GITHUB_REPOSITORY_OWNER": parseOwner(envVars["DRONE_REPO"]),
		"GITHUB_SHA":              envVars["DRONE_COMMIT_SHA"],
		"GITHUB_RUN_ID":           envVars["DRONE_BUILD_NUMBER"],
		"GITHUB_RUN_ATTEMPT":      envVars["DRONE_BUILD_NUMBER"],
		"GITHUB_WORKSPACE":        "/github/workspace",
		"GITHUB_SERVER_URL":       "https://github.com",
		"GITHUB_API_URL":          "https://api.github.com",
		"GITHUB_GRAPHQL_URL":      "https://api.github.com/graphql",
		"RUNNER_OS":               ostype,
		"RUNNER_ARCH":             arch,
		"RUNNER_NAME":             "DRONE HOSTED",
		"RUNNER_TEMP":             "/tmp",
		"RUNNER_TOOL_CACHE":       runner_tool_cache,
		"CI":                      "true",
		"GITHUB_ACTIONS":          "true",
	}, actionEnvVars)

	// Handle tag vs branch for ref name and type
	if tagName := envVars["DRONE_TAG"]; tagName != "" {
		actionEnvVars["GITHUB_REF_NAME"] = tagName
		actionEnvVars["GITHUB_REF_TYPE"] = "tag"
	} else if branchName := envVars["DRONE_BRANCH"]; branchName != "" {
		actionEnvVars["GITHUB_REF_NAME"] = branchName
		actionEnvVars["GITHUB_REF_TYPE"] = "branch"
	}

	secretEnvVars := make(map[string]string)
	for _, secretName := range secrets {
		if os.Getenv(secretName) != "" {
			secretEnvVars[secretName] = os.Getenv(secretName)
		}
	}

	if err := godotenv.Write(actionEnvVars, envFile); err != nil {
		return errors.Wrap(err, "failed to write environment variables file")
	}
	if err := godotenv.Write(secretEnvVars, secretFile); err != nil {
		return errors.Wrap(err, "failed to write secret variables file")
	}
	return nil
}

// Return environment variables set in a map format
func getEnvVars() map[string]string {
	m := make(map[string]string)
	for _, e := range os.Environ() {
		if i := strings.Index(e, "="); i >= 0 {
			m[e[:i]] = e[i+1:]
		}
	}
	return m
}

func Exists(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

// helper function gets the owner from a repository slug.
func parseOwner(s string) (owner string) {
	if parts := strings.Split(s, "/"); len(parts) == 2 {
		return parts[0]
	}
	return
}

func combineEnv(a, b map[string]string) map[string]string {
	c := make(map[string]string, len(a)+len(b))
	for k, v := range a {
		c[k] = v
	}
	for k, v := range b {
		c[k] = v
	}
	return c
}
