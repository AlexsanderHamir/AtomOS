# AtomOS Package Manager

A simple package manager for AtomOS blocks that can download, store, and delete binary blocks from GitHub repositories. It supports automatic installation detection, LSP entry parsing, and cross-platform binary management.

## Features

- **Download and Install**: Download blocks from GitHub repositories
- **Simple Downloads**: Download binaries directly from GitHub releases
- **Clean Uninstall**: Remove blocks and clean up all associated files
- **Metadata Tracking**: Track installation metadata including versions and timestamps
- **Cross-Platform Support**: Support for Linux, macOS, and Windows binaries
- **Existing Installation Support**: Automatically detect and load existing installations
- **Installation Validation**: Validate existing installations for integrity
- **LSP Entry Support**: Parse and store LSP entries from agentic_support.yaml
- **GitHub Token Support**: Support for private repositories using GITHUB_TOKEN
- **Version Management**: Support for versioned metadata storage
- **Testing Support**: Custom test directory support for unit testing

## Missing Features

The following features are mentioned in the original documentation but are not yet implemented:

- **Update Management**: Update installed blocks to newer versions (`Update` method)
- **Installation Statistics**: Get detailed statistics about installed blocks (`GetInstallationStats` method)
- **List Public Method**: Public method to list all installed blocks
- **Get Info Method**: Get information about a specific block by name
- **IsLoaded Method**: Check if the installation has been loaded into memory
- **GetLoadedBlocks Method**: Return all blocks loaded from existing installation
- **IsBlockLoaded Method**: Check if a specific block is loaded in memory

## API Methods

### Core Methods

- `NewPackageManager() *PackageManager` - Creates a new package manager instance using default directories and loads existing installation if present
- `NewPackageManagerWithTestDir(testDir string) *PackageManager` - Creates a new package manager instance with a custom test directory for testing purposes

- `Install(req InstallRequest) (*BlockMetadata, error)` - Installs a block and returns its metadata
- `Uninstall(Blockname string) error` - Removes an installed block
- `list() (*listResult, error)` - Lists all installed blocks (internal method)

### Installation Management Methods

- `isExistingInstallation() bool` - Checks if this is an existing installation
- `GetLoadedBlock(Blockname string) (*BlockMetadata, bool)` - Returns a specific block by name from loaded installation
- `checkBinariesExistAndLoad() error` - Validates the integrity of an existing installation

### Helper Methods

- `fetchBlockInfo(repo string) (*BlockInfo, error)` - Fetches block information from GitHub repository
- `getLatestRelease(repo string) (*GitHubRelease, error)` - Gets the latest release from a GitHub repository
- `downloadBinary(repo, version string, blockInfo *BlockInfo) (string, error)` - Downloads a binary for the current platform
- `getBinaryNameForPlatform(blockInfo *BlockInfo) (string, error)` - Returns the binary name for the current platform
- `storeMetadata(metadata *BlockMetadata) error` - Stores block metadata to disk
- `getMetadata(Blockname string) (*BlockMetadata, error)` - Retrieves block metadata from disk

## Installation Management

The package manager automatically detects and loads existing installations. When you create a new `PackageManager` instance:

1. **New Installation**: If no `.atomos` directory exists, a new installation is created
2. **Existing Installation**: If a `.atomos` directory exists with installed blocks, the installation is loaded into memory
3. **Validation**: Existing installations are validated to ensure all metadata files have corresponding binaries
4. **Caching**: Loaded blocks are cached in memory for faster access

### Loading Behavior

- The package manager checks for existing block directories in the `~/.atomos/` directory
- For each block directory, it looks for metadata files in the `metadata/` subdirectory
- If metadata files exist, it validates that all corresponding binaries are present in the `bin/` subdirectory
- Valid installations are loaded into memory and cached in a map structure for O(1) lookups by block name
- Invalid installations show warnings but allow the package manager to continue working
- The package manager uses the most recently modified metadata file for each block

### Installation Detection

The package manager determines if an installation exists by:

1. Checking if the `.atomos` directory exists
2. Scanning for block directories containing metadata files
3. Validating that binaries exist for all metadata files
4. Loading valid blocks into the `loadedBlocks` map

### Error Handling

- Missing binaries for existing metadata files cause installation validation to fail
- The package manager will show a warning but continue to work for new installations
- Invalid metadata files are skipped during loading

## Usage Example

```go
package main

import (
    "fmt"
    "log"

    packagemanager "github.com/AlexsanderHamir/AtomOS/pkgs/package_manager"
)

func main() {
    // Create a new package manager instance
    pm := packagemanager.NewPackageManager()

    // Install a block
    installReq := packagemanager.InstallRequest{
        Repo:    "AlexsanderHamir/prof",
        Version: "1.8.1",
        Force:   false,
    }

    metadata, err := pm.Install(installReq)
    if err != nil {
        log.Fatalf("Installation failed: %v", err)
    }

    fmt.Printf("Installed block: %s v%s\n", metadata.Name, metadata.Version)
    fmt.Printf("Binary path: %s\n", metadata.BinaryPath)

    // Get the installed block
    block, exists := pm.GetLoadedBlock(metadata.Name)
    if exists {
        fmt.Printf("Block is loaded: %s\n", block.Name)
    }

    // Uninstall the block
    err = pm.Uninstall(metadata.Name)
    if err != nil {
        log.Fatalf("Uninstall failed: %v", err)
    }

    fmt.Println("Block uninstalled successfully")
}
```

## Installation

The package manager is part of the AtomOS project.

## Block Structure

Blocks must have an `agentic_support.yaml` file at the root of their repository with the following structure:

```yaml
name: my-block
description: A description of my block
version: v1.0.0
source:
  type: github
  repo: owner/repo
binary:
  from: release
  assets:
    linux-amd64: my-block-linux-amd64
    darwin-amd64: my-block-darwin-amd64
    darwin-arm64: my-block-darwin-arm64
    windows-amd64: my-block-windows-amd64.exe
lsp:
  entries:
    run:
      name: run
      description: Run the block
      inputs:
        - name: input_file
          type: string
      outputs:
        - name: output_file
          type: string
```

### Schema Details

- **name**: The block name (required)
- **description**: A description of the block (required)
- **version**: The version tag, typically prefixed with 'v' (e.g., "v1.0.0") (required)
- **source**: Source configuration (required)
  - **type**: Must be "github" (required)
  - **repo**: GitHub repository in "owner/repo" format (required)
- **binary**: Binary configuration (required)
  - **from**: Must be "release" (required)
  - **assets**: Platform-specific binary names (required)
    - Supported platforms: `linux-amd64`, `darwin-amd64`, `darwin-arm64`, `windows-amd64`
- **lsp**: LSP (Language Server Protocol) entries configuration (required)
  - **entries**: Map of entry names to entry definitions (required)
    - Each entry must have: `name`, `description`, `inputs`, `outputs`
    - **inputs**: Array of input parameters with `name` and `type`
    - **outputs**: Array of output parameters with `name` and `type`

## Directory Structure

The package manager creates the following directory structure:

```
~/.atomos/
└── block-name/
    ├── bin/
    │   └── binary-file
    └── metadata/
        └── version.json
```

Each block is organized in its own subdirectory containing:

- `bin/`: Contains the executable binary files for the block
- `metadata/`: Contains versioned metadata files (e.g., `1.8.1.json`) with block information

### Installation Directory

The default installation directory is `~/.atomos/` (where `~` is the user's home directory). The package manager uses the following fallback logic to determine the home directory:

1. `os.UserHomeDir()` - Standard Go method
2. `$HOME` environment variable
3. Current working directory (as fallback)
4. System temporary directory (as last resort)

For testing purposes, you can use `NewPackageManagerWithTestDir(testDir string)` to create a package manager instance that uses a custom directory instead of the home directory.

## GitHub Integration

The package manager integrates with GitHub for downloading blocks and binaries:

### Authentication

- **Public Repositories**: No authentication required
- **Private Repositories**: Requires `GITHUB_TOKEN` environment variable
- The token must have appropriate permissions to access the repository and download releases

### Supported Operations

- **Fetch Block Info**: Downloads `agentic_support.yaml` from repository root
- **Get Latest Release**: Fetches the latest release information
- **Download Binaries**: Downloads platform-specific binaries from release assets
- **Version Support**: Supports both tagged releases (with/without 'v' prefix)

### Error Handling

- **404 Not Found**: Repository or file doesn't exist
- **401/403 Unauthorized**: Authentication failed or insufficient permissions
- **Rate Limiting**: GitHub API rate limits are respected
- **Network Errors**: Timeout and connection errors are handled gracefully

## Data Types

### BlockMetadata

Represents metadata about an installed block:

```go
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
```

### InstallRequest

Represents a request to install a block:

```go
type InstallRequest struct {
    Repo    string `json:"repo"`
    Version string `json:"version"`
    Force   bool   `json:"force"` // Force reinstall even if already installed
}
```

### Entry

Represents an LSP entry from the block:

```go
type Entry struct {
    Name        string   `yaml:"name"`
    Description string   `yaml:"description"`
    Inputs      []Input  `yaml:"inputs"`
    Outputs     []Output `yaml:"outputs"`
}
```
