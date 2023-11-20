package utils

import (
	"os"

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
	Uses string            `yaml:"uses"`
	With map[string]string `yaml:"with"`
	Env  map[string]string `yaml:"env"`
}

const (
	workflowEvent = "push"
	workflowName  = "drone-github-action"
	jobName       = "action"
	runsOnImage   = "ubuntu-latest"
)

func CreateWorkflowFile(ymlFile string, action string,
	with map[string]string, env map[string]string) error {
	j := job{
		Name:   jobName,
		RunsOn: runsOnImage,
		Steps: []step{
			{
				Uses: action,
				With: with,
				Env:  env,
			},
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

	if err = os.WriteFile(ymlFile, out, 0644); err != nil {
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
