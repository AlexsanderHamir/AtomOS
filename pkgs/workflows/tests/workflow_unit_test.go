package tests

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/AlexsanderHamir/AtomOS/pkgs/workflows"
	"github.com/joho/godotenv"
)

func TestMain(m *testing.M) {
	err := godotenv.Load("../../../.env")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load .env: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()
	os.Exit(code)
}

func TestCompileWorkflow(t *testing.T) {
	t.Parallel()

	testDir := fmt.Sprintf("./atomos-test-dir-%s", t.Name())
	wm := workflows.NewWorkflowManager(testDir)

	defer func() {
		if err := os.RemoveAll(testDir); err != nil {
			t.Logf("failed to remove test dir: %v", err)
		}
	}()

	t.Run("compile", func(t *testing.T) {
		workflowPath := filepath.Join("validcases", "pipeline_workflow_atoms.yaml")
		err := wm.CompileWorkflow(workflowPath)
		if err != nil {
			t.Fatalf("CompileWorkflow failed: %v", err)
		}
	})

	t.Run("run", func(t *testing.T) {
		err := wm.RunWorkFlow("simple three-block workflow")
		if err != nil {
			t.Fatalf("RunWorkFlow failed: %v", err)
		}
	})
}
