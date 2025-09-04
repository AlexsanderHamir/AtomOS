package workflows

import (
	"fmt"

	packagemanager "github.com/AlexsanderHamir/AtomOS/pkgs/package_manager"
)

type WorkflowManager struct {
	pkgmanager *packagemanager.PackageManager
	workflows  map[string]*CompiledWorkFlow
}

type CompiledWorkFlow struct {
	blocksInfo map[string]*packagemanager.BlockMetadata
}

// NewWorkflowManager creates and returns a new WorkflowManager with a default PackageManager.
func NewWorkflowManager() *WorkflowManager {
	return &WorkflowManager{
		pkgmanager: packagemanager.NewPackageManager(),
		workflows:  map[string]*CompiledWorkFlow{},
	}
}

// TODO - Need to figure out this feature
//
// Compile workflow is responsible for downloading blocks, and ensuring that the connection are correct.
func (wm *WorkflowManager) CompileWorkflow(workflowPath string) error {
	// 1. Parse workflow.
	workflow, err := parseWorkflow(workflowPath)
	if err != nil {
		return err
	}

	// 2. Download blocks.
	blocksInfo := make(map[string]*packagemanager.BlockMetadata)
	for _, blocks := range workflow.Blocks {
		installReq := packagemanager.InstallRequest{
			Repo:    blocks.GitHub,
			Version: blocks.Version,
		}

		blockInfo, err := wm.pkgmanager.Install(installReq)
		if err != nil {
			return fmt.Errorf("couldn't install block, failed: %w", err)
		}

		blocksInfo[blockInfo.Name] = blockInfo
	}

	// 3. Build connections.

	// Store workflow info.
	newWorkflow := &CompiledWorkFlow{
		blocksInfo: blocksInfo,
	}
	wm.workflows[workflow.Name] = newWorkflow

	return nil
}
