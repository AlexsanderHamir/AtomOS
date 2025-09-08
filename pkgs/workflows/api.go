package workflows

import (
	"errors"
	"fmt"

	packagemanager "github.com/AlexsanderHamir/AtomOS/pkgs/package_manager"
	"github.com/dominikbraun/graph"
)

// NewWorkflowManager creates and returns a new WorkflowManager with a default PackageManager.
func NewWorkflowManager(path string) *WorkflowManager {
	return &WorkflowManager{
		pkgmanager: packagemanager.NewPackageManagerWithTestDir(path),
		metadata:   map[Blockname]*packagemanager.BlockMetadata{},
		workflows:  map[Workflowname]graph.Graph[string, *Block]{},
		results:    map[Outputkey]Outputres{},
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

		wm.metadata[Blockname(block.Name)] = blockMetadata
	}

	g := buildGraph(rawWorkflow)
	wm.workflows[Workflowname(rawWorkflow.Name)] = g

	return nil
}

// BFS traversal with connection access
func (wm *WorkflowManager) RunWorkFlow(wfn Workflowname) error {
	g := wm.workflows[wfn]

	startNode := findRootNode(g)
	if startNode == "" {
		return errors.New("no root node found")
	}

	visited := make(map[string]bool)
	queue := []string{startNode}
	level := 0

	adjacencyMap, err := g.AdjacencyMap()
	if err != nil {
		return fmt.Errorf("error getting adjacency map: %v", err)
	}

	for len(queue) > 0 {
		levelSize := len(queue)

		for range levelSize {
			currentNode := queue[0]
			queue = queue[1:]

			if visited[currentNode] {
				continue
			}
			visited[currentNode] = true

			block, err := g.Vertex(currentNode)
			if err != nil {
				return fmt.Errorf("error getting block %s: %v", currentNode, err)
			}

			incomingConnections, incomingFromBlocks := getIncoming(adjacencyMap, currentNode)
			outgoingConnections, outgoingToBlocks := getOutGoing(adjacencyMap, currentNode)

			blockMetadata := wm.metadata[Blockname(block.Name)]
			excArgs := ExecuteArgs{block, blockMetadata, incomingConnections, incomingFromBlocks, outgoingConnections, outgoingToBlocks}

			err = wm.executeBlock(excArgs)
			if err != nil {
				return fmt.Errorf("error executing block %s: %v", block.Name, err)
			}

			for target := range adjacencyMap[currentNode] {
				if !visited[target] {
					queue = append(queue, target)
				}
			}
		}
		fmt.Println()
		level++
	}

	return nil
}

// Execute block with access to all connections
func (wm *WorkflowManager) executeBlock(excArgs ExecuteArgs) error {
	for i, edge := range excArgs.incon {
		fmt.Println(i, edge)
	}

	shouldUseSource := len(excArgs.incon) <= 0
	binary := excArgs.metadata.BinaryPath

	for _, edge := range excArgs.outcon {
		inputpath := edge.Properties.Attributes["input"]
		outputpath := edge.Properties.Attributes["output"]
		fromEntry := edge.Properties.Attributes["fromEntry"]
		sourcePath := edge.Properties.Attributes["source"]

		if shouldUseSource {
			if err := wm.fromSource(binary, fromEntry, outputpath, sourcePath); err != nil {
				return fmt.Errorf("fromSource failed: %w", err)
			}
		}

		if err := wm.fromNode(binary, fromEntry, inputpath, outputpath); err != nil {
			return fmt.Errorf("fromNode failed: %w", err)
		}
	}

	return nil
}
