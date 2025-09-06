package tests

import (
	"path/filepath"
	"testing"

	"github.com/AlexsanderHamir/AtomOS/pkgs/workflows"
)

func TestCompileWorkflow(t *testing.T) {
	workflowPath := filepath.Join("..", "..", "..", "examples", "codereview_workflow_atomos.yaml")

	wm := workflows.NewWorkflowManager()

	err := wm.CompileWorkflow(workflowPath)
	if err != nil {
		t.Fatalf("CompileWorkflow failed: %v", err)
	}

	t.Log("Workflow compiled successfully")
}
