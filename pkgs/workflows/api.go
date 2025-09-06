package workflows

import (
	"fmt"

	packagemanager "github.com/AlexsanderHamir/AtomOS/pkgs/package_manager"
	"github.com/dominikbraun/graph"
)

type blockname string
type workflowname string
type WorkflowManager struct {
	pkgmanager *packagemanager.PackageManager
	metadata   map[blockname]*packagemanager.BlockMetadata
	workflows  map[workflowname]graph.Graph[string, *Block]
}

// NewWorkflowManager creates and returns a new WorkflowManager with a default PackageManager.
func NewWorkflowManager() *WorkflowManager {
	return &WorkflowManager{
		pkgmanager: packagemanager.NewPackageManager(),
		metadata:   map[blockname]*packagemanager.BlockMetadata{},
	}
}

func (wm *WorkflowManager) CompileWorkflow(workflowPath string) error {
	rawWorkflow, err := parseWorkflow(workflowPath)
	if err != nil {
		return fmt.Errorf("parseWorkflow failed: %w", err)
	}

	for _, block := range rawWorkflow.Blocks {
		installReq := packagemanager.InstallRequest{
			Repo:    block.GitHub,
			Version: block.Version,
			Force:   block.Force,
		}

		blockMetadata, err := wm.pkgmanager.Install(installReq)
		if err != nil {
			return fmt.Errorf("failed to install block '%s': %w", block.Name, err)
		}

		wm.metadata[blockname(block.Name)] = blockMetadata
	}

	g := buildGraph(rawWorkflow)
	wm.workflows[workflowname(rawWorkflow.Name)] = g

	return nil
}

// GetMetadata returns the metadata for a specific block
func (wm *WorkflowManager) GetMetadata(blockName string) (*packagemanager.BlockMetadata, bool) {
	metadata, exists := wm.metadata[blockname(blockName)]
	return metadata, exists
}

// GetWorkflow returns the workflow graph for a specific workflow name
func (wm *WorkflowManager) GetWorkflow(workflowName string) (graph.Graph[string, *Block], bool) {
	workflow, exists := wm.workflows[workflowname(workflowName)]
	return workflow, exists
}

// GetAllMetadata returns all stored block metadata
func (wm *WorkflowManager) GetAllMetadata() map[string]*packagemanager.BlockMetadata {
	result := make(map[string]*packagemanager.BlockMetadata)
	for blockName, metadata := range wm.metadata {
		result[string(blockName)] = metadata
	}
	return result
}
