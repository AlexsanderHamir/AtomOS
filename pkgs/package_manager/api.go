package packagemanager

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// NewPackageManager creates a new package manager instance
// If the hidden atoms directory already exists, it will be loaded from that directory
func NewPackageManager() *PackageManager {
	return NewPackageManagerWithTestDir("")
}

// NewPackageManagerWithTestDir creates a new package manager instance with a custom test directory
// If testDir is empty, it uses the default behavior (home directory)
// If testDir is provided, it creates the hidden directory under the test directory for testing purposes
func NewPackageManagerWithTestDir(testDir string) *PackageManager {
	var installDir string

	if testDir != "" {
		// Testing mode: create hidden directory under the provided test directory
		installDir = filepath.Join(testDir, getDefaultInstallDirPathName)
	} else {
		// Normal mode: use default home directory
		installDir = getDefaultInstallDirPath()
	}

	var dirExists bool
	if _, err := os.Stat(installDir); err == nil {
		dirExists = true
	}

	pm := &PackageManager{
		InstallDir:   installDir,
		loadedBlocks: make(map[string]*BlockMetadata),
	}

	if dirExists {
		if err := pm.loadExistingInstallation(); err != nil {
			fmt.Printf("Warning: Failed to load existing installation: %v\n", err)
		}
		return pm
	}

	os.MkdirAll(installDir, 0755)

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
			metadata, metaErr := pm.getMetadata(blockInfo.Name)
			if metaErr != nil {
				return nil, fmt.Errorf("block '%s' is already installed but failed to read metadata: %w", blockInfo.Name, metaErr)
			}
			log.Printf("%s coming from cache", blockInfo.Name)
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
		LSPEntries:  convertEntriesToMap(blockInfo.Entries),
	}

	if err := pm.storeMetadata(metadata); err != nil {
		return nil, fmt.Errorf("failed to store metadata: %w", err)
	}

	pm.loadedBlocks[metadata.Name] = metadata

	return metadata, nil
}

// GetLoadedBlock returns a specific block by name from the loaded installation
func (pm *PackageManager) GetLoadedBlock(blockName string) (*BlockMetadata, bool) {
	if pm.loadedBlocks == nil {
		return nil, false
	}
	block, exists := pm.loadedBlocks[blockName]
	return block, exists
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

	metadataPath := filepath.Join(pm.InstallDir, blockName, "metadata", fmt.Sprintf("%s.json", metadata.Version))
	if err := os.Remove(metadataPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove metadata: %v", err)
	}

	// Attempt to remove block directory if empty
	_ = os.Remove(filepath.Join(pm.InstallDir, blockName))

	// Remove from loaded blocks if the package manager is loaded
	if pm.loadedBlocks != nil {
		delete(pm.loadedBlocks, blockName)
	}

	return nil
}
