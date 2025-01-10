package utils

import (
	"fmt"
	"log"
	"path/filepath"

	// "github.com/joho/godotenv"
	"golang.org/x/exp/slog"
)

func GetOutputVars(codedir, name string) []string {
	outputVars := make([]string, 0)
	_, _, relPath, _, err := parseActionName(name)
	log.Println(relPath)
	if err != nil {
		slog.Warn(fmt.Sprintf("failed to parse action name: %s with err: %v", name, err))
		return outputVars
	}
	actionYmlFilePath := filepath.Join(codedir, relPath)
	spec, err := parseFile(getActionYamlFname(actionYmlFilePath))
	log.Println(spec)
	if err != nil {
		slog.Warn(fmt.Sprintf("failed to parse output vars: %v", err))
	}

	if spec != nil && spec.Outputs != nil {
		for k := range spec.Outputs {
			outputVars = append(outputVars, k)
		}
	}
	return outputVars



	// spec, err := parseFile("/tmp/action.yml")
	// if err != nil {
	// 	slog.Warn(fmt.Sprintf("failed to parse output vars: %v", err))
	// }

	// if spec != nil && spec.Outputs != nil {
	// 	for k := range spec.Outputs {
	// 		outputVars = append(outputVars, k)
	// 	}
	// }
	// return outputVars
}