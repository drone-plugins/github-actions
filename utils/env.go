package utils

import (
	"os"
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
