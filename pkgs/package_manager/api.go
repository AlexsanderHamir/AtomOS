package packagemanager

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

// IsBlockLoaded checks if a specific block is loaded in memory
func (pm *PackageManager) IsBlockLoaded(blockName string) bool {
	if !pm.isLoaded {
		return false
	}
	_, exists := pm.loadedBlocks[blockName]
	return exists
}

// GetInstallationStats returns statistics about the current installation
func (pm *PackageManager) GetInstallationStats() (*InstallationStats, error) {
	stats := &InstallationStats{
		InstallDir: pm.InstallDir,
		CacheDir:   pm.CacheDir,
		IsExisting: pm.isExistingInstallation(),
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
