package plugin

import (
	"fmt"
	"github.com/drone-plugins/drone-github-actions/daemon"
	"github.com/drone-plugins/drone-github-actions/utils"
	"github.com/pkg/errors"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)

const (
	envFile          = "/tmp/action.env"
	secretFile       = "/tmp/action.secrets"
	workflowFile     = "/tmp/workflow.yml"
	eventPayloadFile = "/tmp/event.json"
	// outputFile       = "/tmp/output.env"
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
	repo, ref, ok := utils.ParseLookup(p.Action.Uses)
	sha := ""
	if !ok {
		log.Println("failed to get repo and ref")
	}

	log.Println(p.Action.Uses, repo, ref, sha)
	outputFile := os.Getenv("DRONE_OUTPUT")
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		_, err := os.OpenFile(outputFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}
	}
	data := `commit_sha=abc123def456
build_number=42
artifact_url=https://example.com/artifacts/42
status=success`

err1 := os.WriteFile(outputFile, []byte(data), 0644)
if err1 != nil {
    log.Fatalf("Failed to write to output file: %v", err1)
}
	log.Println(outputFile)
	content, err3 := os.ReadFile(outputFile)
	if err3 != nil {
		log.Fatalf("Failed to read the file: %v", err3)
	}

	fmt.Println("File contents:")
	fmt.Println(string(content))

	outputVar := utils.GetOutputVars("/harness", p.Action.Uses)
	log.Printf("Output Variables: %v\n", outputVar)
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
		"-C",
		"/tmp/engine",
		// "--output-file",
		// outputFile,
		"-b",
		// "--detect-event",
		// "--eventpath",
		// outputFile,
		// "-v",
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
