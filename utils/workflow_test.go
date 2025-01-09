package utils

import (
	"fmt"
	"os"
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
	outputVars := []string{"out1", "out2"}

	// With output variables
	err := CreateWorkflowFile(workflowFile, action, with, env, outputFile, outputVars)
	assert.NoError(t, err)

	content, err := os.ReadFile(workflowFile)
	assert.NoError(t, err)

	// Check general structure
	assert.Contains(t, string(content), "uses: some-action@v1")
	assert.Contains(t, string(content), "steps:")
	assert.Contains(t, string(content), "id: stepIdentifier")

	// Check the `run` command
	assert.Contains(t, string(content), "out1=${{ steps.stepIdentifier.outputs.out1 }}")
	assert.Contains(t, string(content), "out2=${{ steps.stepIdentifier.outputs.out2 }}")
	assert.Contains(t, string(content), fmt.Sprintf("> %s", outputFile))

	// Without output variables
	err = CreateWorkflowFile(workflowFile, action, with, env, outputFile, []string{})
	assert.NoError(t, err)
	content, err = os.ReadFile(workflowFile)
	assert.NoError(t, err)
	assert.NotContains(t, string(content), "run: echo")
}

func TestSetOutputVariables(t *testing.T) {
	outputVars := []string{"var1", "var2"}
	prevStepId := "prevStep"
	outputFile := "/tmp/output"

	// With output variables
	step := setOutputVariables(prevStepId, outputFile, outputVars)
	assert.Equal(t, "output variables", step.Name)
	assert.Contains(t, step.Run, "var1=${{ steps.prevStep.outputs.var1 }}")
	assert.Contains(t, step.Run, "var2=${{ steps.prevStep.outputs.var2 }}")

	// No output variables
	step = setOutputVariables(prevStepId, outputFile, []string{})
	assert.Empty(t, step.Name)
	assert.Empty(t, step.Run)
}

func TestOutputVarWinScript(t *testing.T) {
	outputVars := []string{"var1", "var2"}
	prevStepId := "prevStep"
	outputFile := "C:/output.txt"

	script := outputVarWinScript(outputVars, prevStepId, outputFile)
	assert.Contains(t, script, "out = 'var1=${{ steps.prevStep.outputs.var1 }}\\n'")
	assert.Contains(t, script, "out + 'var2=${{ steps.prevStep.outputs.var2 }}\\n';")
	assert.Contains(t, script, "f = open('C:/output.txt', 'wb'); f.write(bytes(out, 'UTF-8')); f.close()")
}
