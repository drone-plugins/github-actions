package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseActionOutputs(t *testing.T) {
	testDir := t.TempDir()

	// Valid action.yml file
	validAction := `
outputs:
  var1:
    description: "A sample output variable"
  var-2:
    description: "Another variable"`
	err := os.WriteFile(filepath.Join(testDir, "action.yml"), []byte(validAction), 0644)
	assert.NoError(t, err)

	// Debugging file creation
	_, statErr := os.Stat(filepath.Join(testDir, "action.yml"))
	assert.NoError(t, statErr, "action.yml file not created")

	outputs, err := ParseActionOutputs(testDir)
	assert.NoError(t, err)
	assert.ElementsMatch(t, outputs, []string{"var1", "var-2"})

	// Invalid action.yml
	invalidAction := `invalid_yaml`
	err = os.WriteFile(filepath.Join(testDir, "action.yml"), []byte(invalidAction), 0644)
	assert.NoError(t, err)

	_, err = ParseActionOutputs(testDir)
	assert.Error(t, err)

	// No action.yml or action.yaml
	os.Remove(filepath.Join(testDir, "action.yml"))
	outputs, err = ParseActionOutputs(testDir)
	assert.NoError(t, err)
	assert.Empty(t, outputs)
}
