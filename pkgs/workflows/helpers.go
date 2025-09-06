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

	for _, connection := range rwf.Connections {
		g.AddEdge(connection.FromBlock, connection.ToBlock,
			graph.EdgeAttribute("fromEntry", connection.FromEntry),
			graph.EdgeAttribute("output", connection.Output),
			graph.EdgeAttribute("toEntry", connection.ToEntry),
			graph.EdgeAttribute("input", connection.Input),
		)
	}

	return g
}
