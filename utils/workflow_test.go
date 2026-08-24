package utils

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateWorkflowFile(t *testing.T) {
	testDir := t.TempDir()
	workflowFile := testDir + "/workflow.yml"
	outputFile := testDir + "/output"

	action := "some-action@v1"
	with := map[string]string{"input1": "value1"}
	env := map[string]string{"VAR": "value"}
	outputVars := []string{"out1", "out-2"}

	// With output variables
	err := CreateWorkflowFile(workflowFile, action, with, env, outputFile, outputVars)
	assert.NoError(t, err)

	content, err := os.ReadFile(workflowFile)
	assert.NoError(t, err)

	// Check general structure
	assert.Contains(t, string(content), "uses: some-action@v1")
	assert.Contains(t, string(content), "steps:")
	assert.Contains(t, string(content), "id: stepIdentifier")

	// Check the run command references each output and appends to the file
	assert.Contains(t, string(content), "steps.stepIdentifier.outputs.out1")
	assert.Contains(t, string(content), "steps.stepIdentifier.outputs.out-2")
	assert.Contains(t, string(content), fmt.Sprintf(">> %s", outputFile))
	assert.Contains(t, string(content), "__B64__")

	// Without output variables
	err = CreateWorkflowFile(workflowFile, action, with, env, outputFile, []string{})
	assert.NoError(t, err)
	content, err = os.ReadFile(workflowFile)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "name: output variables")
	assert.Contains(t, string(content), fmt.Sprintf(": > %s", outputFile))
	assert.Contains(t, string(content), "if: \"false\"")
}

// TestSetOutputVariables_MultilineValueRoundTrip is the CI-24318 fix verification.
//
// It renders setOutputVariables, substitutes a multiline value the way act
// would, runs the real bash, and parses the resulting file with the same
// contract Harness's delegate reader uses (key=__B64__<payload> lines with
// automatic base64 decode). The full multiline value must survive.
func TestSetOutputVariables_MultilineValueRoundTrip(t *testing.T) {
	testDir := t.TempDir()
	outputFile := filepath.Join(testDir, "output.env")

	prevStepId := "stepIdentifier"
	outputVars := []string{"changelog", "new_tag"}

	step := setOutputVariables(prevStepId, outputFile, outputVars)

	multilineChangelog := "* commit A\n* commit B\n* commit C"
	newTag := "v1.2.3"

	shellCmd := substituteAll(step.Run, prevStepId, map[string]string{
		"changelog": multilineChangelog,
		"new_tag":   newTag,
	})

	t.Logf("shell command the plugin would run:\n%s", shellCmd)

	cmd := exec.Command("bash", "-c", shellCmd)
	out, err := cmd.CombinedOutput()
	assert.NoError(t, err, "bash exec failed: %s", string(out))

	raw, err := os.ReadFile(outputFile)
	assert.NoError(t, err)
	t.Logf("resulting DRONE_OUTPUT file:\n---\n%s\n---", string(raw))

	parsed := parseB64Output(t, outputFile)
	assert.Equal(t, multilineChangelog, parsed["changelog"],
		"multiline changelog must round-trip intact")
	assert.Equal(t, newTag, parsed["new_tag"],
		"single-line value must round-trip intact")
}

// TestSetOutputVariables_MultilineValuePreservesBlankLines confirms that
// empty lines inside a multiline value are preserved.
func TestSetOutputVariables_MultilineValuePreservesBlankLines(t *testing.T) {
	testDir := t.TempDir()
	outputFile := filepath.Join(testDir, "output.env")

	prevStepId := "s"
	outputVars := []string{"body"}

	step := setOutputVariables(prevStepId, outputFile, outputVars)

	val := "line 1\n\nline 3\n\n\nline 6"
	shellCmd := substituteAll(step.Run, prevStepId, map[string]string{"body": val})

	cmd := exec.Command("bash", "-c", shellCmd)
	out, err := cmd.CombinedOutput()
	assert.NoError(t, err, "bash exec failed: %s", string(out))

	parsed := parseB64Output(t, outputFile)
	assert.Equal(t, val, parsed["body"])
}

// TestSetOutputVariables_ValuesWithSpecialChars covers characters that
// broke the old writer: double quotes, dollar signs, backticks, and
// backslashes. All must survive because the payload is base64-encoded.
func TestSetOutputVariables_ValuesWithSpecialChars(t *testing.T) {
	testDir := t.TempDir()
	outputFile := filepath.Join(testDir, "output.env")

	prevStepId := "s"
	outputVars := []string{"tricky"}

	step := setOutputVariables(prevStepId, outputFile, outputVars)

	trickyVal := "quote:\" dollar:$USER backtick:`id` backslash:\\ end"
	shellCmd := substituteAll(step.Run, prevStepId, map[string]string{"tricky": trickyVal})

	t.Logf("shell command:\n%s", shellCmd)

	cmd := exec.Command("bash", "-c", shellCmd)
	out, err := cmd.CombinedOutput()
	assert.NoError(t, err, "bash exec failed: %s", string(out))

	parsed := parseB64Output(t, outputFile)
	assert.Equal(t, trickyVal, parsed["tricky"])
}

func TestSetOutputVariables(t *testing.T) {
	outputVars := []string{"var1", "var2"}
	prevStepId := "prevStep"
	outputFile := "/tmp/output"

	// With output variables: one per-variable block that appends __B64__ line.
	step := setOutputVariables(prevStepId, outputFile, outputVars)
	assert.Equal(t, "output variables", step.Name)
	assert.Equal(t, "true", step.If)
	assert.Contains(t, step.Run, "steps.prevStep.outputs.var1")
	assert.Contains(t, step.Run, "steps.prevStep.outputs.var2")
	assert.Contains(t, step.Run, fmt.Sprintf(">> %s", outputFile))
	assert.Contains(t, step.Run, "__B64__")
	// Each variable gets its own heredoc marker for ingestion.
	assert.Regexp(t, regexp.MustCompile(`<<'HARNESS_EOF_[0-9a-f]{16}'`), step.Run)
	assert.Contains(t, step.Run, `echo "var1=__B64__$__b64"`)
	assert.Contains(t, step.Run, `echo "var2=__B64__$__b64"`)

	// No output variables: file is truncated, step is gated off.
	step = setOutputVariables(prevStepId, outputFile, []string{})
	assert.Equal(t, "output variables", step.Name)
	assert.Contains(t, step.Run, ": > /tmp/output")
	assert.Equal(t, "false", step.If)
}

// substituteAll replaces every ${{ steps.<id>.outputs.<name> }} placeholder
// in shellCmd with its corresponding value, mimicking act's substitution.
func substituteAll(shellCmd, prevStepId string, values map[string]string) string {
	out := shellCmd
	for name, val := range values {
		placeholder := fmt.Sprintf("${{ steps.%s.outputs.%s }}", prevStepId, name)
		out = strings.ReplaceAll(out, placeholder, val)
	}
	return out
}

// parseB64Output reads a DRONE_OUTPUT file in the same shape Harness's
// delegate reader does: parse key=value lines, then base64-decode any value
// starting with the __B64__ prefix.
func parseB64Output(t *testing.T, path string) map[string]string {
	t.Helper()
	f, err := os.Open(path)
	assert.NoError(t, err)
	defer f.Close()

	parsed := map[string]string{}
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		idx := strings.Index(line, "=")
		if idx <= 0 {
			continue
		}
		key := line[:idx]
		val := line[idx+1:]
		if strings.HasPrefix(val, "__B64__") {
			raw, err := base64.StdEncoding.DecodeString(val[len("__B64__"):])
			assert.NoError(t, err)
			parsed[key] = string(raw)
			continue
		}
		parsed[key] = val
	}
	assert.NoError(t, scanner.Err())
	return parsed
}
