package utils

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type workflow struct {
	Name string         `yaml:"name"`
	On   string         `yaml:"on"`
	Jobs map[string]job `yaml:"jobs"`
}

type job struct {
	Name   string `yaml:"name"`
	RunsOn string `yaml:"runs-on"`
	Steps  []step `yaml:"steps"`
}

type step struct {
	Id    string            `yaml:"id,omitempty"`
	Name  string            `yaml:"name,omitempty"`
	Uses  string            `yaml:"uses,omitempty"`
	Run   string            `yaml:"run,omitempty"`
	With  map[string]string `yaml:"with,omitempty"`
	Env   map[string]string `yaml:"env,omitempty"`
	Shell string            `yaml:"shell,omitempty"`
	If    string            `yaml:"if,omitempty"`
}

const (
	stepId        = "stepIdentifier"
	workflowEvent = "push"
	workflowName  = "drone-github-action"
	jobName       = "action"
	runsOnImage   = "ubuntu-latest"
)

func CreateWorkflowFile(ymlFile string, action string,
	with map[string]string, env map[string]string, outputFile string, outputVars []string) error {
	steps := []step{
		{
			Id:   stepId,
			Uses: action,
			With: with,
			Env:  env,
		},
	}
	// Only append the output step when there are outputs. Emitting a dummy
	// run step with empty uses: "" is invalid GHA YAML and fails act >= 0.2.89.
	if outputStep, ok := setOutputVariables(stepId, outputFile, outputVars); ok {
		steps = append(steps, outputStep)
	}

	j := job{
		Name:   jobName,
		RunsOn: runsOnImage,
		Steps:  steps,
	}
	wf := &workflow{
		Name: workflowName,
		On:   getWorkflowEvent(),
		Jobs: map[string]job{
			jobName: j,
		},
	}

	out, err := yaml.Marshal(&wf)
	if err != nil {
		return errors.Wrap(err, "failed to create action workflow yml")
	}

	if err = ioutil.WriteFile(ymlFile, out, 0644); err != nil {
		return errors.Wrap(err, "failed to write yml workflow file")
	}

	return nil
}

func getWorkflowEvent() string {
	buildEvent := os.Getenv("DRONE_BUILD_EVENT")
	if buildEvent == "push" || buildEvent == "pull_request" || buildEvent == "tag" {
		return buildEvent
	}
	return "custom"
}

func setOutputVariables(prevStepId, outputFile string, outputVars []string) (step, bool) {
	if len(outputFile) == 0 || len(outputVars) == 0 {
		logrus.Infof("No output variables detected in action.yml; skipping output file generation.")
		return step{}, false
	}

	cmd := ""
	for _, outputVar := range outputVars {
		cmd += fmt.Sprintf("%s=${{ steps.%s.outputs.%s }}\n", outputVar, prevStepId, outputVar)
	}

	cmd = fmt.Sprintf("echo \"%s\" > %s", cmd, outputFile)
	return step{
		Name: "output variables",
		Run:  cmd,
	}, true
}
