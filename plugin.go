package plugin

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/drone-plugins/drone-github-actions/daemon"
	"github.com/drone-plugins/drone-github-actions/utils"
	"github.com/drone/plugin/cloner"
	"github.com/drone/plugin/plugin/github"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	envFile          = "/tmp/action.env"
	secretFile       = "/tmp/action.secrets"
	workflowFile     = "/tmp/workflow.yml"
	eventPayloadFile = "/tmp/event.json"
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

	ctx := context.Background()
	repoURL, ref, ok := github.ParseLookup(p.Action.Uses)
	if !ok {
		logrus.Warnf("Invalid 'uses' format: %s", p.Action.Uses)
		return fmt.Errorf("invalid 'uses' format: %s", p.Action.Uses)
	}
	logrus.Infof("Parsed 'uses' string. Repo: %s, Ref: %s", repoURL, ref)

	// Clone the GH Action repository using `cloner` with parsed repo and ref
	clone := cloner.NewCache(cloner.NewDefault())
	codedir, cloneErr := clone.Clone(ctx, repoURL, ref, "")
	if cloneErr != nil {
		logrus.Warnf("Failed to clone GH Action: %v", cloneErr)
	} else {
		logrus.Infof("Successfully cloned GH Action to %s", codedir)
	}

	outputFile := os.Getenv("DRONE_OUTPUT")
	outputVars := []string{}

	if codedir != "" {
		var err error
		outputVars, err = utils.ParseActionOutputs(codedir)
		if err != nil {
			logrus.Warnf("Could not parse action.yml outputs from %s: %v", codedir, err)
		}
	}

	if len(outputVars) == 0 {
		logrus.Infof("No outputs were found in action.yml for repo: %s", repoURL)
	}

	if err := utils.CreateWorkflowFile(workflowFile, p.Action.Uses,
		p.Action.With, p.Action.Env, outputFile, outputVars); err != nil {
		return err
	}

	if err := utils.CreateEnvAndSecretFile(envFile, secretFile, secrets); err != nil {
		return err
	}

	outputFilePath := GetDirPath(outputFile)
	containerOptions := fmt.Sprintf("-v=%s:%s", outputFilePath, outputFilePath)

	cmdArgs := []string{
		"-W",
		workflowFile,
		"-P",
		fmt.Sprintf("ubuntu-latest=%s", p.Action.Image),
		"--secret-file",
		secretFile,
		"--env-file",
		envFile,
		"-b",
		"--detect-event",
		"--container-options",
		fmt.Sprintf("\"%s\"", containerOptions),
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

func GetDirPath(filePath string) string {
	return filepath.Dir(filePath)
}
