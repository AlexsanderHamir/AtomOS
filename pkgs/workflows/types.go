package workflows

import (
	packagemanager "github.com/AlexsanderHamir/AtomOS/pkgs/package_manager"
	"github.com/dominikbraun/graph"
)

// Workflow represents the top-level workflow definition parsed from YAML.
// It includes metadata, a list of blocks, and the connections between them.
type RawWorkflow struct {
	Name        string       `yaml:"workflow_name"`
	Version     string       `yaml:"version"`
	Description string       `yaml:"description"`
	Blocks      []Block      `yaml:"blocks"`
	Connections []Connection `yaml:"connections"`
}

// Block describes a reusable component in the workflow that can expose entries.
type Block struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
	GitHub  string `yaml:"github"`
	Force   bool   `yaml:"force"`
}

// Connection wires outputs from one block entry to inputs of another block entry.
type Connection struct {
	FromBlock string `yaml:"from_block"`
	FromEntry string `yaml:"from_entry"`
	Output    string `yaml:"output"`
	ToBlock   string `yaml:"to_block"`
	ToEntry   string `yaml:"to_entry"`
	Input     string `yaml:"input"`
	Source    string `yaml:"source"`
}

type blockname string
type workflowname string

type WorkflowManager struct {
	pkgmanager *packagemanager.PackageManager
	metadata   map[blockname]*packagemanager.BlockMetadata
	workflows  map[workflowname]graph.Graph[string, *Block]
}

type ExecuteArgs struct {
	block    *Block
	metadata *packagemanager.BlockMetadata
	incon    []graph.Edge[string]
	inblock  []string
	outcon   []graph.Edge[string]
	outblock []string
}
