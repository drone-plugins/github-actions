package utils

import (
	"fmt"
	// "path/filepath"

	// "github.com/joho/godotenv"
	"golang.org/x/exp/slog"
)

func GetOutputVars() []string {
	outputVars := make([]string, 0)
	spec, err := parseFile("/tmp/action.yml")
	if err != nil {
		slog.Warn(fmt.Sprintf("failed to parse output vars: %v", err))
	}

	if spec != nil && spec.Outputs != nil {
		for k := range spec.Outputs {
			outputVars = append(outputVars, k)
		}
	}
	return outputVars
}