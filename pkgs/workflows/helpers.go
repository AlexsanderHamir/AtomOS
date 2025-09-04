package workflows

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// parseWorkflow reads a YAML workflow file from the provided path and returns a Workflow struct.
func parseWorkflow(path string) (*Workflow, error) {
	fileBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read workflow file: %w", err)
	}

	var wf Workflow
	if err := yaml.Unmarshal(fileBytes, &wf); err != nil {
		return nil, fmt.Errorf("unmarshal workflow yaml: %w", err)
	}

	return &wf, nil
}
