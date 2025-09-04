package packagemanager

import (
	"fmt"
	"os"
	"path/filepath"
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

// Install downloads a block and returns its metadata
func (pm *PackageManager) Install(req InstallRequest) (*BlockMetadata, error) {
	blockInfo, err := pm.fetchBlockInfo(req.Repo)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch block info: %w", err)
	}

	if !req.Force {
		if pm.isBlockInstalled(blockInfo.Name) {
			// Return existing metadata if already installed
			metadata, metaErr := pm.getMetadata(blockInfo.Name)
			if metaErr != nil {
				return nil, fmt.Errorf("block '%s' is already installed but failed to read metadata: %w", blockInfo.Name, metaErr)
			}
			return metadata, nil
		}
	}

	version := req.Version
	if version == "" {
		latestRelease, err := pm.getLatestRelease(req.Repo)
		if err != nil {
			return nil, fmt.Errorf("failed to get latest release: %w", err)
		}
		version = latestRelease.TagName
	}

	binaryPath, err := pm.downloadBinary(req.Repo, version, blockInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to download binary: %w", err)
	}

	metadata := &BlockMetadata{
		Name:        blockInfo.Name,
		Version:     version,
		SourceRepo:  req.Repo,
		BinaryPath:  binaryPath,
		InstalledAt: time.Now(),
		LastUpdated: time.Now(),
		IsActive:    true,
		LSPEntries:  blockInfo.LSP.Entries,
	}

	if err := pm.storeMetadata(metadata); err != nil {
		return nil, fmt.Errorf("failed to store metadata: %w", err)
	}

	return metadata, nil
}

// GetLoadedBlock returns a specific block by name from the loaded installation
func (pm *PackageManager) GetLoadedBlock(blockName string) (BlockMetadata, bool) {
	if !pm.isLoaded {
		return BlockMetadata{}, false
	}
	block, exists := pm.loadedBlocks[blockName]
	return block, exists
}

// Update updates an installed block to a newer version
func (pm *PackageManager) Update(req UpdateRequest) (*UpdateResult, error) {
	metadata, err := pm.getMetadata(req.BlockName)
	if err != nil {
		return &UpdateResult{
			Success: false,
			Message: fmt.Sprintf("Block '%s' is not installed: %v", req.BlockName, err),
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
			Message: fmt.Sprintf("Block '%s' is already at version %s", req.BlockName, version),
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

	binaryPath, err := pm.downloadBinary(metadata.SourceRepo, version, blockInfo)
	if err != nil {
		return &UpdateResult{
			Success: false,
			Message: fmt.Sprintf("Failed to download binary: %v", err),
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
	metadata.LastUpdated = time.Now()
	metadata.LSPEntries = blockInfo.LSP.Entries

	if err := pm.storeMetadata(metadata); err != nil {
		return &UpdateResult{
			Success: false,
			Message: fmt.Sprintf("Failed to update metadata: %v", err),
		}, err
	}

	return &UpdateResult{
		Success:    true,
		Message:    fmt.Sprintf("Successfully updated block '%s' from %s to %s", req.BlockName, oldVersion, version),
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

	if err := os.Remove(metadata.BinaryPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove binary: %v", err)
	}

	metadataPath := filepath.Join(pm.InstallDir, "metadata", blockName, fmt.Sprintf("%s.json", metadata.Version))
	if err := os.Remove(metadataPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove metadata: %v", err)
	}

	// Attempt to remove block directory if empty
	_ = os.Remove(filepath.Join(pm.InstallDir, "metadata", blockName))

	return nil
}
