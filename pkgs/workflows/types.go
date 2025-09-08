// Copyright (c) 2025 Alexsander Hamir Gomes Baptista
//
// This file is part of AtomOS and licensed under the Sustainable Use License (SUL).
// You may use, modify, and redistribute this software for personal or internal business use.
// Offering it as a commercial hosted service requires a separate license.
//
// Full license: see the LICENSE file in the root of this repository
// or contact alexsanderhamirgomesbaptista@gmail.com.

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
	Input     string `yaml:"input"`
	Source    string `yaml:"source"`
}

type Blockname string
type Workflowname string
type Outputkey string
type Outputres string

type WorkflowManager struct {
	pkgmanager *packagemanager.PackageManager
	metadata   map[Blockname]*packagemanager.BlockMetadata
	workflows  map[Workflowname]graph.Graph[string, *Block]
	results    map[Outputkey]Outputres
}

type ExecuteArgs struct {
	block    *Block
	metadata *packagemanager.BlockMetadata
	incon    []graph.Edge[string]
	inblock  []string
	outcon   []graph.Edge[string]
	outblock []string
}
