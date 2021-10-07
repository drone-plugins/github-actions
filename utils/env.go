package utils

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/pkg/errors"
)

func CreateEnvAndSecretFile(envFile, secretFile string, secrets []string) error {
	envVars := getEnvVars()

	envBuf := ""
	for key, val := range envVars {
		if !strings.HasPrefix(key, "PLUGIN_") && !Exists(secrets, key) {
			envBuf += fmt.Sprintf("%s=%s\n", key, val)
		}
	}

	secretBuf := ""
	for _, secretName := range secrets {
		if os.Getenv(secretName) != "" {
			secretBuf += fmt.Sprintf("%s=%s\n", secretName, os.Getenv(secretName))
		}
	}

	if err := ioutil.WriteFile(envFile, []byte(envBuf), 0644); err != nil {
		return errors.Wrap(err, "failed to write environment variables file")
	}

	if err := ioutil.WriteFile(secretFile, []byte(secretBuf), 0644); err != nil {
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
