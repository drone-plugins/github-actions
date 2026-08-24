package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

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
	Uses  string            `yaml:"uses"`
	Run   string            `yaml:"run,omitempty"`
	With  map[string]string `yaml:"with"`
	Env   map[string]string `yaml:"env"`
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
	j := job{
		Name:   jobName,
		RunsOn: runsOnImage,
		Steps: []step{
			{
				Id:   stepId,
				Uses: action,
				With: with,
				Env:  env,
			},
			setOutputVariables(stepId, outputFile, outputVars),
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

// newEOFMarker returns a per-run heredoc terminator that is extremely unlikely
// to collide with any output value content. Falls back to a static suffix if
// the crypto reader fails (should be effectively impossible on real systems).
func newEOFMarker() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return "HARNESS_EOF_FALLBACK"
	}
	return "HARNESS_EOF_" + hex.EncodeToString(buf)
}

// setOutputVariables generates a workflow step that appends each declared
// action output to $DRONE_OUTPUT as a single line "key=__B64__<payload>".
// The Harness delegate's output-var reader (harness/godotenv) detects the
// __B64__ prefix and base64-decodes the payload transparently, so multiline
// values and values containing any special characters round-trip cleanly.
//
// The raw value is ingested via a quoted heredoc so bash does not perform
// variable expansion or escape processing on it.
func setOutputVariables(prevStepId, outputFile string, outputVars []string) step {
	skip := len(outputFile) == 0 || len(outputVars) == 0
	if skip {
		logrus.Infof("No output variables detected in action.yml; skipping output file generation.")
		return step{
			Name: "output variables",
			Run:  fmt.Sprintf(": > %s", outputFile),
			If:   "false",
		}
	}

	marker := newEOFMarker()
	var b strings.Builder
	// Truncate the file first so re-runs do not append onto stale content.
	fmt.Fprintf(&b, ": > %s\n", outputFile)
	for _, outputVar := range outputVars {
		// Read the substituted value into __val via a quoted heredoc kept at
		// the top level (not inside $(...)) to sidestep a bash 3.2 parser bug
		// with double-quotes inside `$(cat <<'EOF' ... EOF)`. The heredoc
		// terminator is generated per plugin run so it will not collide with
		// value content. IFS= and -r prevent whitespace or backslash mangling;
		// -d '' reads until NUL (never present), so read consumes the full
		// heredoc and always returns non-zero, hence `|| true`.
		fmt.Fprintf(&b, `IFS= read -r -d '' __val <<'%s' || true
${{ steps.%s.outputs.%s }}
%s
__val="${__val%%$'\n'}"
__b64=$(printf '%%s' "$__val" | base64 | tr -d '\n')
echo "%s=__B64__$__b64" >> %s
`, marker, prevStepId, outputVar, marker, outputVar, outputFile)
	}

	return step{
		Name: "output variables",
		Run:  b.String(),
		If:   "true",
	}
}
