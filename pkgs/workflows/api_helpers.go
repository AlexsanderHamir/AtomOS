package workflows

import (
	"fmt"
	"os"

	"github.com/dominikbraun/graph"
	"gopkg.in/yaml.v3"
)

func parseWorkflow(path string) (*RawWorkflow, error) {
	fileBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read workflow file: %w", err)
	}

	var rwf RawWorkflow
	if err := yaml.Unmarshal(fileBytes, &rwf); err != nil {
		return nil, fmt.Errorf("unmarshal workflow yaml: %w", err)
	}

	return &rwf, nil
}

func buildGraph(rwf *RawWorkflow) graph.Graph[string, *Block] {
	blockHash := func(b *Block) string {
		return b.Name
	}

	g := graph.New(blockHash, graph.Directed(), graph.Acyclic())
	for _, block := range rwf.Blocks {
		g.AddVertex(&block)
	}

	// Infer edges by matching outputs to inputs across connections.
	// For each connection A that produces an output, find every connection B whose
	// input matches A's output. Create an edge from A.FromBlock -> B.FromBlock and
	// carry relevant attributes for execution.
	for _, src := range rwf.Connections {
		for _, dst := range rwf.Connections {
			if dst.Input == "" {
				continue
			}
			if src.Output == "" {
				continue
			}
			if src.Output != dst.Input {
				continue
			}

			g.AddEdge(src.FromBlock, dst.FromBlock,
				graph.EdgeAttribute("fromEntry", src.FromEntry),
				graph.EdgeAttribute("output", src.Output),
				graph.EdgeAttribute("input", dst.Input),
				graph.EdgeAttribute("source", src.Source),
			)
		}
	}

	return g
}

func findRootNode(g graph.Graph[string, *Block]) string {
	adjacencyMap, err := g.AdjacencyMap()
	if err != nil {
		return ""
	}

	hasIncomingEdge := make(map[string]bool)
	for _, targets := range adjacencyMap {
		for target := range targets {
			hasIncomingEdge[target] = true
		}
	}

	for nodeID := range adjacencyMap {
		if !hasIncomingEdge[nodeID] {
			return nodeID
		}
	}

	return ""
}

func getIncoming(adjacencyMap map[string]map[string]graph.Edge[string], currentNode string) ([]graph.Edge[string], []string) {
	// Get incoming connections
	var incomingConnections []graph.Edge[string]
	var incomingFromBlocks []string
	for sourceNode, targets := range adjacencyMap {
		if edge, exists := targets[currentNode]; exists {
			incomingConnections = append(incomingConnections, edge)
			incomingFromBlocks = append(incomingFromBlocks, sourceNode)
		}
	}

	return incomingConnections, incomingFromBlocks
}

func getOutGoing(adjacencyMap map[string]map[string]graph.Edge[string], currentNode string) ([]graph.Edge[string], []string) {
	var outgoingConnections []graph.Edge[string]
	var outgoingToBlocks []string
	if targets, exists := adjacencyMap[currentNode]; exists {
		for targetNode, edge := range targets {
			outgoingConnections = append(outgoingConnections, edge)
			outgoingToBlocks = append(outgoingToBlocks, targetNode)
		}
	}

	return outgoingConnections, outgoingToBlocks
}

// TODO: Both fromSource and fromNode are not completed, we're passing raw data
// without any commands.
func (wm *WorkflowManager) fromSource(binary, entry, outputpath, sourcePath string) error {
	output, err := runBinaryWithPipe(binary, entry, sourcePath)
	if err != nil {
		return fmt.Errorf("running binary failed: %w", err)
	}

	wm.results[Outputkey(outputpath)] = Outputres(output)
	return nil
}

func (wm *WorkflowManager) fromNode(binary, entry, inputPath, outputpath string) error {
	input := wm.results[Outputkey(inputPath)]

	output, err := runBinaryWithString(binary, entry, input)
	if err != nil {
		return fmt.Errorf("running binary with string failed: %w", err)
	}

	wm.results[Outputkey(outputpath)] = Outputres(output)
	return nil
}
