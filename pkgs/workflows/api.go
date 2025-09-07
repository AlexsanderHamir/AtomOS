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
		metadata:   map[blockname]*packagemanager.BlockMetadata{},
		workflows:  map[workflowname]graph.Graph[string, *Block]{},
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

// BFS traversal with connection access
func (wm *WorkflowManager) RunWorkFlow(wfn workflowname) error {
	g := wm.workflows[wfn]

	startNode := findRootNode(g)
	if startNode == "" {
		return errors.New("no root node found")
	}

	fmt.Println("=== BFS TRAVERSAL ===")
	fmt.Printf("Starting from: %s\n", startNode)

	visited := make(map[string]bool)
	queue := []string{startNode}
	level := 0

	adjacencyMap, err := g.AdjacencyMap()
	if err != nil {
		return fmt.Errorf("error getting adjacency map: %v", err)
	}

	for len(queue) > 0 {
		levelSize := len(queue)
		fmt.Printf("Level %d: ", level)

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

			fmt.Printf("%s ", block.Name)

			incomingConnections, incomingFromBlocks := getIncoming(adjacencyMap, currentNode)
			outgoingConnections, outgoingToBlocks := getOutGoing(adjacencyMap, currentNode)

			blockMetadata := wm.metadata[blockname(block.Name)]
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

	fmt.Printf("\n  Executing: %s\n", excArgs.block.Name)

	// Show incoming connections
	fmt.Printf("    Inputs (%d):\n", len(excArgs.incon))
	for i, edge := range excArgs.incon {
		fmt.Printf("      %s -> %s (%s->%s)\n",
			excArgs.inblock[i], excArgs.block.Name,
			edge.Properties.Attributes["output"],
			edge.Properties.Attributes["input"])
	}

	// Show outgoing connections
	fmt.Printf("    Outputs (%d):\n", len(excArgs.outcon))
	for i, edge := range excArgs.outcon {
		fmt.Printf("      %s -> %s (%s->%s)\n",
			excArgs.block.Name, excArgs.outblock[i],
			edge.Properties.Attributes["output"],
			edge.Properties.Attributes["input"])
	}

	// Your block execution logic here
	// You now have access to:
	// - block: the current block
	// - metadata: block metadata
	// - incomingConnections: edges coming in
	// - incomingFromBlocks: source block names
	// - outgoingConnections: edges going out
	// - outgoingToBlocks: target block names

	return nil
}
