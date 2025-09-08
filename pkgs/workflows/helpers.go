package workflows

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

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

	for _, connection := range rwf.Connections {
		g.AddEdge(connection.FromBlock, connection.ToBlock,
			graph.EdgeAttribute("fromEntry", connection.FromEntry),
			graph.EdgeAttribute("output", connection.Output),
			graph.EdgeAttribute("toEntry", connection.ToEntry),
			graph.EdgeAttribute("input", connection.Input),
			graph.EdgeAttribute("source", connection.Source),
		)
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

func runBinaryWithPipe(binary, filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	cmd := exec.Command(binary)
	cmd.Stdin = file

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run the command
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("binary failed: %v, stderr: %s", err, stderr.String())
	}

	return stdout.String(), nil
}

// runBinaryWithString pipes the given input string into the binary's stdin
// and returns the binary's stdout output.
func runBinaryWithString(binary string, input Outputres) (string, error) {
	// Prepare the command
	cmd := exec.Command(binary)

	// Pipe string into stdin
	cmd.Stdin = strings.NewReader(string(input))

	// Capture stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run the command
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("binary failed: %v, stderr: %s", err, stderr.String())
	}

	return stdout.String(), nil
}

// TODO: Both fromSource and fromNode are not completed, we're passing raw data
// without any commands.
func (wm *WorkflowManager) fromSource(binary, outputpath, inputPath string) error {
	output, err := runBinaryWithPipe(binary, inputPath)
	if err != nil {
		return fmt.Errorf("running binary failed: %w", err)
	}

	wm.results[Outputkey(outputpath)] = Outputres(output)
	return nil
}

func (wm *WorkflowManager) fromNode(binary, inputPath, outputpath string) error {
	input := wm.results[Outputkey(inputPath)]

	output, err := runBinaryWithString(binary, input)
	if err != nil {
		return fmt.Errorf("running binary with string failed: %w", err)
	}

	wm.results[Outputkey(outputpath)] = Outputres(output)
	return nil
}
