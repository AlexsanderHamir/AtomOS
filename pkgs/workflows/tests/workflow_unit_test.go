package tests

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/AlexsanderHamir/AtomOS/pkgs/workflows"
)

func TestCompileWorkflow(t *testing.T) {
	t.Parallel()

	testDir := fmt.Sprintf("./atomos-test-dir-%s", t.Name())
	wm := workflows.NewWorkflowManager(testDir)

	t.Run("compile", func(t *testing.T) {
		workflowPath := filepath.Join("validcases", "pipeline_workflow_atoms.yaml")
		err := wm.CompileWorkflow(workflowPath)
		if err != nil {
			t.Fatalf("CompileWorkflow failed: %v", err)
		}
	})

	t.Run("run", func(t *testing.T) {

	})
}
