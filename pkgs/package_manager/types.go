// Copyright (c) 2025 Alexsander Hamir Gomes Baptista
//
// This file is part of AtomOS and licensed under the Sustainable Use License (SUL).
// You may use, modify, and redistribute this software for personal or internal business use.
// Offering it as a commercial hosted service requires a separate license.
//
// Full license: see the LICENSE file in the root of this repository
// or contact alexsanderhamirgomesbaptista@gmail.com.

package packagemanager

import (
	"time"
)

// BlockMetadata represents metadata about an installed block
type BlockMetadata struct {
	Name        string           `json:"name"`
	Version     string           `json:"version"`
	SourceRepo  string           `json:"source_repo"`
	BinaryPath  string           `json:"binary_path"`
	InstalledAt time.Time        `json:"installed_at"`
	LastUpdated time.Time        `json:"last_updated"`
	IsActive    bool             `json:"is_active"`
	LSPEntries  map[string]Entry `json:"lsp_entries,omitempty"`
}

// InstallRequest represents a request to install a block
type InstallRequest struct {
	Repo    string `json:"repo"`
	Version string `json:"version"`
	Force   bool   `json:"force"` // Force reinstall even if already installed
}

// UpdateRequest represents a request to update a block
type UpdateRequest struct {
	Blockname string `json:"block_name"`
	Version   string `json:"version"` // If empty, will check for latest
}

// PackageManager handles block installation, updates, and management
type PackageManager struct {
	InstallDir string
	// Loaded state from existing installation
	loadedBlocks map[string]*BlockMetadata // Cached map of installed blocks by name
}

// BlockInfo represents the information from agentic_support.yaml
type BlockInfo struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Version     string `yaml:"version"`
	Source      struct {
		Type string `yaml:"type"`
		Repo string `yaml:"repo"`
	} `yaml:"source"`
	Binary struct {
		From   string            `yaml:"from"`
		Assets map[string]string `yaml:"assets"`
	} `yaml:"binary"`
	Entries    []Entry `yaml:"entries"`
	BinaryPath string  // Path to the downloaded binary
}

// Entry represents a CLI entry from the block
type Entry struct {
	Name        string   `yaml:"name"`
	Command     string   `yaml:"command"`
	Description string   `yaml:"description"`
	Inputs      []Input  `yaml:"inputs"`
	Outputs     []Output `yaml:"outputs"`
}

// Input represents an input parameter for an entry
type Input struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"`
}

// Output represents an output from an entry
type Output struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"`
}

// GitHubRelease represents a GitHub release with assets
type GitHubRelease struct {
	TagName     string         `json:"tag_name"`
	Name        string         `json:"name"`
	Body        string         `json:"body"`
	Assets      []ReleaseAsset `json:"assets"`
	CreatedAt   string         `json:"created_at"`
	PublishedAt string         `json:"published_at"`
}

// ReleaseAsset represents an asset in a GitHub release
type ReleaseAsset struct {
	ID            int    `json:"id"` // Add this field - it's required!
	Name          string `json:"name"`
	ContentType   string `json:"content_type"`
	Size          int    `json:"size"`
	DownloadURL   string `json:"browser_download_url"`
	DownloadCount int    `json:"download_count"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

// InstallResult represents the result of an installation
type InstallResult struct {
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	BinaryPath string `json:"binary_path,omitempty"`
	Blockname  string `json:"block_name,omitempty"`
	Version    string `json:"version,omitempty"`
}

// UpdateResult represents the result of an update
type UpdateResult struct {
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	OldVersion string `json:"old_version,omitempty"`
	NewVersion string `json:"new_version,omitempty"`
	BinaryPath string `json:"binary_path,omitempty"`
}

// listResult represents the result of listing installed blocks
type listResult struct {
	Blocks []BlockMetadata `json:"blocks"`
	Total  int             `json:"total"`
}

// InstallationStats represents statistics about the package manager installation
type InstallationStats struct {
	InstallDir      string          `json:"install_dir"`
	IsExisting      bool            `json:"is_existing"`
	TotalBlocks     int             `json:"total_blocks"`
	TotalBinarySize int64           `json:"total_binary_size"`
	InstalledBlocks []BlockMetadata `json:"installed_blocks,omitempty"`
}
