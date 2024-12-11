package utils

import (
	"io/ioutil"
	"os"
	"fmt"
	"runtime"
	"github.com/pkg/errors"
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
	Uses  string            `yaml:"uses,omitempty"`
	Name  string            `yaml:"name,omitempty"`
	With  map[string]string `yaml:"with,omitempty"`
	Env   map[string]string `yaml:"env,omitempty"`
	Run   string            `yaml:"run,omitempty"`
	Shell string            `yaml:"shell,omitempty"`
	If    string            `yaml:"if,omitempty"`
}

const (
	workflowEvent = "push"
	workflowName  = "drone-github-action"
	jobName       = "action"
	runsOnImage   = "ubuntu-latest"
	stepId        = "stepId"
)

func CreateWorkflowFile(ymlFile string, action string,
	with map[string]string, env map[string]string, outputFile string, outputVars []string) error {
	j := job{
		Name:   jobName,
		RunsOn: runsOnImage,
		Steps: []step{
			{
				Uses: action,
				With: with,
				Env:  env,
			},
			getOutputVariables(stepId, outputFile, outputVars),
		},
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

func getOutputVariables(prevStepId, outputFile string, outputVars []string) step {
    skip := len(outputFile) == 0 || len(outputVars) == 0
    cmd := ""
    for _, outputVar := range outputVars {
        cmd += fmt.Sprintf("%s=${{ steps.%s.outputs.%s }}\n", outputVar, prevStepId, outputVar)
    }

    if runtime.GOOS == "windows" {
        // Windows: Create the file and write the output variables using Python
        cmd = fmt.Sprintf("python -c \"%s\"", outputVarWinScript(outputVars, outputFile))
    } else {
        // Unix-like: Use `touch` to create the file, then append the output variables
        cmd = fmt.Sprintf("echo \"%s\" > %s",cmd, outputFile)
    }

    s := step{
        Name: "output variables",
        Run:  cmd,
        If:   fmt.Sprintf("%t", !skip),
    }

    if runtime.GOOS == "windows" {
        s.Shell = "powershell"
    }

    return s
}


func outputVarWinScript(outputVars []string, outputFile string) string {
	script := ""
	for idx, outputVar := range outputVars {
		prefix := "out = "
		if idx > 0 {
			prefix += "out + "
		}
		script += fmt.Sprintf("%s'%s=${{ steps.%s.outputs.%s }}\\n';", prefix, outputVar, stepId, outputVar)
	}
	script += fmt.Sprintf("f = open('%s', 'wb'); f.write(bytes(out, 'UTF-8')); f.close()", outputFile)
	return script
}