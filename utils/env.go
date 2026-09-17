package utils

import (
	"os"
	"regexp"
	"strings"

	"github.com/joho/godotenv"
	"github.com/pkg/errors"
)

// validEnvName matches names that are legal in a dotenv file. Harness injects
// a step's output variables into later steps in the stage (e.g. cache-hit and
// node-version from actions/setup-node), and act >= 0.2.89 fails to parse
// hyphenated names in --env-file.
var validEnvName = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

func CreateEnvAndSecretFile(envFile, secretFile string, secrets []string) error {
	envVars := getEnvVars()

	actionEnvVars := make(map[string]string)
	for key, val := range envVars {
		if !validEnvName.MatchString(key) {
			continue
		}
		if !strings.HasPrefix(key, "PLUGIN_") && !Exists(secrets, key) {
			actionEnvVars[key] = val
		}
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
