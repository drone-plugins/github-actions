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

func TestParseLookup(t *testing.T) {
	tests := []struct {
		name string
		uses string
		repo string
		ref  string
		path string
		ok   bool
	}{
		{
			name: "root action",
			uses: "mathieudutour/github-tag-action@v6.2",
			repo: "https://github.com/mathieudutour/github-tag-action",
			ref:  "v6.2",
			path: "",
			ok:   true,
		},
		{
			name: "nested action path",
			uses: "my-corp/my-up2-action-external-management/.github/actions/mathieudutour/github-tag-action@v1",
			repo: "https://github.com/my-corp/my-up2-action-external-management",
			ref:  "v1",
			path: ".github/actions/mathieudutour/github-tag-action",
			ok:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, ref, path, ok := ParseLookup(tt.uses)
			assert.Equal(t, tt.ok, ok)
			assert.Equal(t, tt.repo, repo)
			assert.Equal(t, tt.ref, ref)
			assert.Equal(t, tt.path, path)
		})
	}
}

func TestActionDir(t *testing.T) {
	clone := t.TempDir()

	dir, err := ActionDir(clone, "")
	assert.NoError(t, err)
	assert.Equal(t, clone, dir)

	dir, err = ActionDir(clone, ".github/actions/foo")
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(clone, ".github/actions/foo"), dir)

	_, err = ActionDir(clone, "../outside")
	assert.Error(t, err)

	// Traversal that only escapes after filepath.Clean
	_, err = ActionDir(clone, "foo/../../etc")
	assert.Error(t, err)

	_, err = ActionDir(clone, "foo/../..")
	assert.Error(t, err)

	_, err = ActionDir(clone, "foo/bar/../../../outside")
	assert.Error(t, err)

	// Clean keeps the result inside cloneDir — should succeed
	dir, err = ActionDir(clone, "foo/../.github/actions/bar")
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(clone, ".github/actions/bar"), dir)
}

func TestParseActionOutputsNested(t *testing.T) {
	clone := t.TempDir()
	nested := filepath.Join(clone, ".github", "actions", "tag")
	assert.NoError(t, os.MkdirAll(nested, 0755))

	content := `
outputs:
  new_tag:
    description: "Generated tag"
  new_version:
    description: "Generated version"
`
	assert.NoError(t, os.WriteFile(filepath.Join(nested, "action.yml"), []byte(content), 0644))

	actionDir, err := ActionDir(clone, ".github/actions/tag")
	assert.NoError(t, err)

	outputs, err := ParseActionOutputs(actionDir)
	assert.NoError(t, err)
	assert.ElementsMatch(t, outputs, []string{"new_tag", "new_version"})

	// Root still empty when only nested action.yml exists
	outputs, err = ParseActionOutputs(clone)
	assert.NoError(t, err)
	assert.Empty(t, outputs)
}

func TestParseActionOutputsMissingNested(t *testing.T) {
	clone := t.TempDir()
	actionDir, err := ActionDir(clone, ".github/actions/missing")
	assert.NoError(t, err)

	outputs, err := ParseActionOutputs(actionDir)
	assert.NoError(t, err)
	assert.Empty(t, outputs)
}
