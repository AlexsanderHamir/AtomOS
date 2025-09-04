package packagemanager

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// NewPackageManager creates a new package manager instance
// If the hidden atoms directory already exists, it will be loaded from that directory
func NewPackageManager() *PackageManager {
	installDir := getDefaultInstallDirPath()
	cacheDir := getDefaultCacheDirPath(installDir)

	var dirExists bool
	if _, err := os.Stat(installDir); err == nil {
		dirExists = true
	}

	pm := &PackageManager{
		InstallDir: installDir,
		CacheDir:   cacheDir,
	}

	if dirExists {
		if err := pm.loadExistingInstallation(); err != nil {
			fmt.Printf("Warning: Failed to load existing installation: %v\n", err)
		}
		return pm
	}

	os.MkdirAll(installDir, 0755)
	os.MkdirAll(cacheDir, 0755)

	return pm
}

// IsExistingInstallation checks if this package manager is working with an existing installation
func (pm *PackageManager) IsExistingInstallation() bool {
	if pm.isLoaded {
		return len(pm.loadedBlocks) > 0
	}

	metadataDir := filepath.Join(pm.InstallDir, "metadata")
	if _, err := os.Stat(metadataDir); err != nil {
		return false
	}

	files, err := os.ReadDir(metadataDir)
	if err != nil {
		return false
	}

	// If there are any .json files in the metadata directory, it's an existing installation
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			return true
		}
	}

	return false
}

// GetLoadedBlocks returns the blocks that were loaded from the existing installation
func (pm *PackageManager) GetLoadedBlocks() []BlockMetadata {
	if !pm.isLoaded {
		return nil
	}

	// Convert map to slice
	blocks := make([]BlockMetadata, 0, len(pm.loadedBlocks))
	for _, block := range pm.loadedBlocks {
		blocks = append(blocks, block)
	}
	return blocks
}

// GetLoadedBlock returns a specific block by name from the loaded installation
func (pm *PackageManager) GetLoadedBlock(blockName string) (BlockMetadata, bool) {
	if !pm.isLoaded {
		return BlockMetadata{}, false
	}
	block, exists := pm.loadedBlocks[blockName]
	return block, exists
}

// IsBlockLoaded checks if a specific block is loaded in memory
func (pm *PackageManager) IsBlockLoaded(blockName string) bool {
	if !pm.isLoaded {
		return false
	}
	_, exists := pm.loadedBlocks[blockName]
	return exists
}

// IsLoaded returns whether the installation has been loaded
func (pm *PackageManager) IsLoaded() bool {
	return pm.isLoaded
}

// GetInstallationStats returns statistics about the current installation
func (pm *PackageManager) GetInstallationStats() (*InstallationStats, error) {
	stats := &InstallationStats{
		InstallDir: pm.InstallDir,
		CacheDir:   pm.CacheDir,
		IsExisting: pm.IsExistingInstallation(),
	}

	if stats.IsExisting {
		// Use loaded blocks if available, otherwise fetch from disk
		if pm.isLoaded {
			stats.TotalBlocks = len(pm.loadedBlocks)
			// Convert map to slice for stats
			blocks := make([]BlockMetadata, 0, len(pm.loadedBlocks))
			for _, block := range pm.loadedBlocks {
				blocks = append(blocks, block)
			}
			stats.InstalledBlocks = blocks
		} else {
			// Get list of installed blocks from disk
			listResult, err := pm.List()
			if err != nil {
				return nil, fmt.Errorf("failed to get installation stats: %w", err)
			}
			stats.TotalBlocks = len(listResult.Blocks)
			stats.InstalledBlocks = listResult.Blocks
		}

		// Calculate total size of binaries
		var totalSize int64
		blocks := stats.InstalledBlocks
		for _, block := range blocks {
			if info, err := os.Stat(block.BinaryPath); err == nil {
				totalSize += info.Size()
			}
		}
		stats.TotalBinarySize = totalSize
	}

	return stats, nil
}

// Install downloads and installs a block with SHA256 verification
func (pm *PackageManager) Install(req InstallRequest) (*InstallResult, error) {
	// Fetch block information from the repository
	blockInfo, err := pm.fetchBlockInfo(req.Repo)
	if err != nil {
		return &InstallResult{
			Success: false,
			Message: fmt.Sprintf("Failed to fetch block info: %v", err),
		}, err
	}

	// Check if already installed (unless force is true)
	if !req.Force {
		if pm.isBlockInstalled(blockInfo.Name) {
			return &InstallResult{
				Success: false,
				Message: fmt.Sprintf("Block '%s' is already installed. Use --force to reinstall.", blockInfo.Name),
			}, nil
		}
	}

	// Get the latest release if version is not specified
	version := req.Version
	if version == "" {
		latestRelease, err := pm.getLatestRelease(req.Repo)
		if err != nil {
			return &InstallResult{
				Message: fmt.Sprintf("Failed to get latest release: %v", err),
			}, err
		}
		version = latestRelease.TagName
	}

	// Download and verify the binary
	binaryPath, sha256Hash, err := pm.downloadAndVerifyBinary(req.Repo, version, blockInfo)
	if err != nil {
		return &InstallResult{
			Success: false,
			Message: fmt.Sprintf("Failed to download and verify binary: %v", err),
		}, err
	}

	// Store metadata
	metadata := &BlockMetadata{
		Name:        blockInfo.Name,
		Version:     version,
		SourceRepo:  req.Repo,
		BinaryPath:  binaryPath,
		SHA256:      sha256Hash,
		InstalledAt: time.Now(),
		LastUpdated: time.Now(),
		IsActive:    true,
	}

	if err := pm.storeMetadata(metadata); err != nil {
		return &InstallResult{
			Success: false,
			Message: fmt.Sprintf("Failed to store metadata: %v", err),
		}, err
	}

	return &InstallResult{
		Success:    true,
		Message:    fmt.Sprintf("Successfully installed block '%s' version %s", blockInfo.Name, version),
		BinaryPath: binaryPath,
		BlockName:  blockInfo.Name,
		Version:    version,
	}, nil
}

// Update updates an installed block to a newer version
func (pm *PackageManager) Update(req UpdateRequest) (*UpdateResult, error) {
	// Get current metadata
	metadata, err := pm.getMetadata(req.Name)
	if err != nil {
		return &UpdateResult{
			Success: false,
			Message: fmt.Sprintf("Block '%s' is not installed: %v", req.Name, err),
		}, nil
	}

	// Get the latest release if version is not specified
	version := req.Version
	if version == "" {
		latestRelease, err := pm.getLatestRelease(metadata.SourceRepo)
		if err != nil {
			return &UpdateResult{
				Success: false,
				Message: fmt.Sprintf("Failed to get latest release: %v", err),
			}, err
		}
		version = latestRelease.TagName
	}

	// Check if we're already at the requested version
	if version == metadata.Version {
		return &UpdateResult{
			Success: false,
			Message: fmt.Sprintf("Block '%s' is already at version %s", req.Name, version),
		}, nil
	}

	// Fetch block info for the new version
	blockInfo, err := pm.fetchBlockInfo(metadata.SourceRepo)
	if err != nil {
		return &UpdateResult{
			Success: false,
			Message: fmt.Sprintf("Failed to fetch block info: %v", err),
		}, err
	}

	// Download and verify the new binary
	binaryPath, sha256Hash, err := pm.downloadAndVerifyBinary(metadata.SourceRepo, version, blockInfo)
	if err != nil {
		return &UpdateResult{
			Success: false,
			Message: fmt.Sprintf("Failed to download and verify binary: %v", err),
		}, err
	}

	// Remove old binary
	if err := os.Remove(metadata.BinaryPath); err != nil && !os.IsNotExist(err) {
		return &UpdateResult{
			Success: false,
			Message: fmt.Sprintf("Failed to remove old binary: %v", err),
		}, err
	}

	// Update metadata
	oldVersion := metadata.Version
	metadata.Version = version
	metadata.BinaryPath = binaryPath
	metadata.SHA256 = sha256Hash
	metadata.LastUpdated = time.Now()

	if err := pm.storeMetadata(metadata); err != nil {
		return &UpdateResult{
			Success: false,
			Message: fmt.Sprintf("Failed to update metadata: %v", err),
		}, err
	}

	return &UpdateResult{
		Success:    true,
		Message:    fmt.Sprintf("Successfully updated block '%s' from %s to %s", req.Name, oldVersion, version),
		OldVersion: oldVersion,
		NewVersion: version,
		BinaryPath: binaryPath,
	}, nil
}

// Uninstall removes an installed block
func (pm *PackageManager) Uninstall(blockName string) error {
	metadata, err := pm.getMetadata(blockName)
	if err != nil {
		return fmt.Errorf("block '%s' is not installed: %v", blockName, err)
	}

	// Remove binary
	if err := os.Remove(metadata.BinaryPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove binary: %v", err)
	}

	// Remove metadata
	metadataPath := filepath.Join(pm.InstallDir, "metadata", fmt.Sprintf("%s.json", blockName))
	if err := os.Remove(metadataPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove metadata: %v", err)
	}

	return nil
}

// List returns all installed blocks
func (pm *PackageManager) List() (*ListResult, error) {
	metadataDir := filepath.Join(pm.InstallDir, "metadata")
	if err := os.MkdirAll(metadataDir, 0755); err != nil {
		return nil, err
	}

	files, err := os.ReadDir(metadataDir)
	if err != nil {
		return nil, err
	}

	var blocks []BlockMetadata
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			blockName := strings.TrimSuffix(file.Name(), ".json")
			metadata, err := pm.getMetadata(blockName)
			if err != nil {
				continue
			}
			blocks = append(blocks, *metadata)
		}
	}

	return &ListResult{
		Blocks: blocks,
		Total:  len(blocks),
	}, nil
}

// GetInfo returns information about a specific installed block
func (pm *PackageManager) GetInfo(blockName string) (*BlockMetadata, error) {
	return pm.getMetadata(blockName)
}
