package plugin

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"log"
	"github.com/drone-plugins/drone-github-actions/daemon"
	"github.com/drone-plugins/drone-github-actions/utils"
	"github.com/pkg/errors"
)

const (
	envFile          = "/tmp/action.env"
	secretFile       = "/tmp/action.secrets"
	workflowFile     = "/tmp/workflow.yml"
	eventPayloadFile = "/tmp/event.json"
	outputFile       = "/tmp/action.yml"
)

var (
	secrets = []string{"GITHUB_TOKEN"}
)

type (
	Action struct {
		Uses         string
		With         map[string]string
		Env          map[string]string
		Image        string
		EventPayload string // Webhook event payload
		Actor        string
		Verbose      bool
	}

	Plugin struct {
		Action Action
		Daemon daemon.Daemon // Docker daemon configuration
	}
)

// Exec executes the plugin step
func (p Plugin) Exec() error {
	if err := daemon.StartDaemon(p.Daemon); err != nil {
		return err
	}
	log.Println(p.Action.Uses)
	outputVar := utils.GetOutputVars("/root/.cache/act",p.Action.Uses)
	log.Println(outputVar)
	name := p.Action.Uses
	if err := utils.CreateWorkflowFile(workflowFile, name,
		p.Action.With, p.Action.Env, outputFile, outputVar); err != nil {
		return err
	}

	if err := utils.CreateEnvAndSecretFile(envFile, secretFile, secrets); err != nil {
		return err
	}

	cmdArgs := []string{
		"-W",
		workflowFile,
		"-P",
		fmt.Sprintf("ubuntu-latest=%s", p.Action.Image),
		"--secret-file",
		secretFile,
		"--env-file",
		envFile,
		// "--output-file",
		// outputFile,
		"-b",
		"--detect-event",
	}

	// optional arguments
	if p.Action.Actor != "" {
		cmdArgs = append(cmdArgs, "--actor")
		cmdArgs = append(cmdArgs, p.Action.Actor)
	}

	if p.Action.EventPayload != "" {
		if err := ioutil.WriteFile(eventPayloadFile, []byte(p.Action.EventPayload), 0644); err != nil {
			return errors.Wrap(err, "failed to write event payload to file")
		}

		cmdArgs = append(cmdArgs, "--eventpath", eventPayloadFile)
	}

	if p.Action.Verbose {
		cmdArgs = append(cmdArgs, "-v")
	}

	cmd := exec.Command("act", cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	trace(cmd)

	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

// trace writes each command to stdout with the command wrapped in an xml
// tag so that it can be extracted and displayed in the logs.
func trace(cmd *exec.Cmd) {
	fmt.Fprintf(os.Stdout, "+ %s\n", strings.Join(cmd.Args, " "))
}
