# AtomOS Package Manager

A simple package manager for AtomOS blocks that can download, store, delete, and update binary blocks from GitHub repositories.

## Features

- **Download and Install**: Download blocks from GitHub repositories
- **Simple Downloads**: Download binaries directly from GitHub releases
- **Update Management**: Update installed blocks to newer versions
- **Clean Uninstall**: Remove blocks and clean up all associated files
- **Metadata Tracking**: Track installation metadata including versions and timestamps
- **Cross-Platform Support**: Support for Linux, macOS, and Windows binaries
- **Existing Installation Support**: Automatically detect and load existing installations
- **Installation Validation**: Validate existing installations for integrity
- **Installation Statistics**: Get detailed statistics about installed blocks

## API Methods

### Core Methods

– `NewPackageManager() *PackageManager` - Creates a new package manager instance using default directories and loads existing installation if present

- `Install(req InstallRequest) (*InstallResult, error)` - Installs a block
- `Update(req UpdateRequest) (*UpdateResult, error)` - Updates an installed block
- `Uninstall(blockName string) error` - Removes an installed block
- `List() (*ListResult, error)` - Lists all installed blocks
- `GetInfo(blockName string) (*BlockMetadata, error)` - Gets information about a specific block

### Installation Management Methods

- `isExistingInstallation() bool` - Checks if this is an existing installation
- `IsLoaded() bool` - Checks if the installation has been loaded into memory
- `GetLoadedBlocks() []BlockMetadata` - Returns the blocks loaded from existing installation
- `GetLoadedBlock(blockName string) (BlockMetadata, bool)` - Returns a specific block by name from loaded installation
- `IsBlockLoaded(blockName string) bool` - Checks if a specific block is loaded in memory
- `checkBinariesExistAndLoad() error` - Validates the integrity of an existing installation
- `GetInstallationStats() (*InstallationStats, error)` - Gets detailed installation statistics

## Installation Management

The package manager automatically detects and loads existing installations. When you create a new `PackageManager` instance:

1. **New Installation**: If no `.atomos` directory exists, a new installation is created
2. **Existing Installation**: If a `.atomos` directory exists with installed blocks, the installation is loaded into memory
3. **Validation**: Existing installations are validated to ensure all metadata files have corresponding binaries
4. **Caching**: Loaded blocks are cached in memory for faster access

### Loading Behavior

- The package manager checks for existing metadata files in the `~/.atomos/metadata/` directory
- If metadata files exist, it validates that all corresponding binaries are present
- Valid installations are loaded into memory and marked as "loaded"
- Blocks are cached in a map structure for O(1) lookups by block name
- Invalid installations show warnings but allow the package manager to continue working

## Installation

The package manager is part of the AtomOS project.

## Block Structure

Blocks must have an `agentic_support.yaml` file at the root of their repository with the following structure:

```yaml
name: my-block
description: A description of my block
version: 1.0.0
source:
  type: github
  repo: owner/repo
binary:
  from: releases
  assets:
    linux-amd64: my-block-linux-amd64
    darwin-amd64: my-block-darwin-amd64
    windows-amd64: my-block-windows-amd64.exe
entries:
  - name: run
    command: ./my-block
    description: Run the block
    inputs:
      - name: input_file
        type: string
    outputs:
      - name: output_file
        type: string
```

## Directory Structure

The package manager creates the following directory structure:

```
~/.atomos/
│   └── block-name/
│       └── binary-file
~/.atomos/
├── metadata/
│   └── block-name.json
└── cache/
    └── (temporary files)
```

## API Usage

```go
package main

import (
    "fmt"
    "log"

    "github.com/AlexsanderHamir/AtomOS/pkgs/package_manager"
)

func main() {
    // Create package manager instance
    // Uses default directories: ~/.atomos and ~/.atomos/cache
    // If ~/.atomos already exists, it will be loaded and validated
    pm := package_manager.NewPackageManager()

        // Check if this is an existing installation
    if pm.isExistingInstallation() {
        fmt.Println("Loading existing AtomOS installation...")

        // Check if the installation was successfully loaded
        if pm.IsLoaded() {
            fmt.Println("Installation loaded successfully!")

                        // Get the loaded blocks
            loadedBlocks := pm.GetLoadedBlocks()
            fmt.Printf("Found %d loaded blocks\n", len(loadedBlocks))

            for _, block := range loadedBlocks {
                fmt.Printf("- %s (v%s)\n", block.Name, block.Version)
            }

            // Use fast map-based lookups
            if pm.IsBlockLoaded("my-block") {
                block, _ := pm.GetLoadedBlock("my-block")
                fmt.Printf("Found block: %s version %s\n", block.Name, block.Version)
            }
        } else {
            fmt.Println("Installation exists but failed to load")
        }

        // Validate the existing installation
        if err := pm.checkBinariesExistAndLoad(); err != nil {
            fmt.Printf("Warning: Installation validation failed: %v\n", err)
        }

        // Get installation statistics
        stats, err := pm.GetInstallationStats()
        if err != nil {
            log.Fatalf("Failed to get installation stats: %v", err)
        }

        fmt.Printf("Found %d installed blocks (total size: %d bytes)\n",
            stats.TotalBlocks, stats.TotalBinarySize)
    } else {
        fmt.Println("Creating new AtomOS installation...")
    }

    // Install a block
    req := package_manager.InstallRequest{
        Repo:    "owner/repo",
        Version: "v1.0.0",
        Force:   false,
    }

    result, err := pm.Install(req)
    if err != nil {
        log.Fatalf("Install failed: %v", err)
    }

    fmt.Printf("Install result: %+v\n", result)

    // List installed blocks
    listResult, err := pm.List()
    if err != nil {
        log.Fatalf("List failed: %v", err)
    }

    fmt.Printf("Installed blocks: %d\n", listResult.Total)
    for _, block := range listResult.Blocks {
        fmt.Printf("- %s (v%s)\n", block.Name, block.Version)
    }
}
```

## Security

- Binaries are stored in isolated directories per block
- Metadata is stored separately from binaries for easy cleanup

## Error Handling

The package manager provides detailed error messages for common issues:

- Repository not found
- Binary not available for current platform
- Network connectivity issues
- Permission errors

## Testing

Run the tests:

```bash
go test ./pkgs/package_manager/...
```
