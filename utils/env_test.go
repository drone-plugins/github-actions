package utils

import (
	"path/filepath"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

func TestValidEnvName(t *testing.T) {
	valid := []string{"NORMAL_VAR", "_leading_underscore", "GITHUB_TOKEN", "path2"}
	for _, name := range valid {
		assert.True(t, validEnvName.MatchString(name), "expected %q to be valid", name)
	}

	// Hyphenated names come from action outputs that Harness injects into
	// later steps, e.g. cache-hit and node-version from actions/setup-node.
	invalid := []string{"cache-hit", "node-version", "2leading_digit", "has space", ""}
	for _, name := range invalid {
		assert.False(t, validEnvName.MatchString(name), "expected %q to be invalid", name)
	}
}

func TestCreateEnvAndSecretFileSkipsInvalidNames(t *testing.T) {
	testDir := t.TempDir()
	envFile := filepath.Join(testDir, "action.env")
	secretFile := filepath.Join(testDir, "action.secrets")

	t.Setenv("HARNESS_GHA_TEST", "kept")
	t.Setenv("cache-hit", "dropped")
	t.Setenv("node-version", "v16.20.2")
	t.Setenv("PLUGIN_USES", "dropped")
	t.Setenv("GITHUB_TOKEN", "token")

	err := CreateEnvAndSecretFile(envFile, secretFile, []string{"GITHUB_TOKEN"})
	assert.NoError(t, err)

	// The env file must be parseable, which is what act >= 0.2.89 requires.
	env, err := godotenv.Read(envFile)
	assert.NoError(t, err)
	assert.Equal(t, "kept", env["HARNESS_GHA_TEST"])
	assert.NotContains(t, env, "cache-hit")
	assert.NotContains(t, env, "node-version")
	assert.NotContains(t, env, "PLUGIN_USES")
	assert.NotContains(t, env, "GITHUB_TOKEN")

	secretsEnv, err := godotenv.Read(secretFile)
	assert.NoError(t, err)
	assert.Equal(t, "token", secretsEnv["GITHUB_TOKEN"])
}
