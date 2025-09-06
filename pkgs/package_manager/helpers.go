package packagemanager

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"gopkg.in/yaml.v3"
)

// fetchBlockInfo fetches the agentic_support.yaml file from the repository
func (pm *PackageManager) fetchBlockInfo(repo string) (*BlockInfo, error) {
	rawURL := fmt.Sprintf("https://raw.githubusercontent.com/%s/main/agentic_support.yaml", repo)

	// Try main branch first
	resp, err := http.Get(rawURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch agentic_support.yaml from main branch: %w", err)
	}
	defer resp.Body.Close()

	// If main branch doesn't exist, try master branch
	if resp.StatusCode == http.StatusNotFound {
		rawURL = fmt.Sprintf("https://raw.githubusercontent.com/%s/master/agentic_support.yaml", repo)
		resp, err = http.Get(rawURL)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch agentic_support.yaml from master branch: %w", err)
		}
		defer resp.Body.Close()
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch agentic_support.yaml: HTTP %d", resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var blockInfo BlockInfo
	if err := yaml.Unmarshal(content, &blockInfo); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return &blockInfo, nil
}

// convertEntriesToMap converts a slice of Entry to a map[string]Entry using the entry name as the key
func convertEntriesToMap(entries []Entry) map[string]Entry {
	result := make(map[string]Entry)
	for _, entry := range entries {
		result[entry.Name] = entry
	}
	return result
}

// getLatestRelease fetches the latest release from GitHub
func (pm *PackageManager) getLatestRelease(repo string) (*GitHubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch latest release: HTTP %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to decode release JSON: %w", err)
	}

	return &release, nil
}

// getReleaseByTag fetches a specific GitHub release by tag and is tolerant
// to tags with or without a leading 'v'.
func (pm *PackageManager) getReleaseByTag(repo, tag string) (*GitHubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/tags/v%s", repo, tag)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch release by tag: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var release GitHubRelease
		if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
			return nil, fmt.Errorf("failed to decode release JSON: %w", err)
		}
		return &release, nil
	}

	if resp.StatusCode != http.StatusNotFound {
		return nil, fmt.Errorf("failed to fetch release by tag '%s': HTTP %d", tag, resp.StatusCode)
	}

	return nil, fmt.Errorf("release not found for tag '%s' (tried with/without 'v')", tag)
}

// downloadBinary downloads a binary for the current platform
func (pm *PackageManager) downloadBinary(repo, version string, blockInfo *BlockInfo) (string, error) {
	osName := runtime.GOOS
	arch := runtime.GOARCH

	// Map Go's runtime values to the YAML keys
	platformKey := fmt.Sprintf("%s-%s", osName, arch)

	// Handle Windows executable extension
	if osName == "windows" {
		platformKey = fmt.Sprintf("%s-%s", osName, arch)
	}

	binaryName, exists := blockInfo.Binary.Assets[platformKey]
	if !exists {
		return "", fmt.Errorf("no binary found for platform %s", platformKey)
	}

	// Create per-block bin directory under the package manager's install directory
	binDir := filepath.Join(pm.InstallDir, blockInfo.Name, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create bin directory: %w", err)
	}

	// Construct the local file path
	localPath := filepath.Join(binDir, binaryName)

	// Resolve the release and pick the correct asset URL rather than constructing it
	release, err := pm.getReleaseByTag(repo, version)
	if err != nil {
		return "", fmt.Errorf("failed to resolve release '%s': %w", version, err)
	}

	var assetURL string
	for _, a := range release.Assets {
		if a.Name == binaryName {
			assetURL = a.DownloadURL
			break
		}
	}
	if assetURL == "" {
		return "", fmt.Errorf("asset '%s' not found in release %s", binaryName, release.TagName)
	}

	resp, err := http.Get(assetURL)
	if err != nil {
		return "", fmt.Errorf("failed to download binary: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download binary: HTTP %d", resp.StatusCode)
	}

	// Create the local file
	file, err := os.Create(localPath)
	if err != nil {
		return "", fmt.Errorf("failed to create local file: %w", err)
	}
	defer file.Close()

	// Copy the downloaded content to the file
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to write binary to file: %w", err)
	}

	// Make the binary executable on Unix-like systems
	if osName != "windows" {
		if err := os.Chmod(localPath, 0755); err != nil {
			return "", fmt.Errorf("failed to make binary executable: %w", err)
		}
	}

	return localPath, nil
}

// isBlockInstalled checks if a block is already installed
func (pm *PackageManager) isBlockInstalled(blockName string) bool {
	// Consider installed if there's at least one versioned metadata file under <block>/metadata/
	blockDir := filepath.Join(pm.InstallDir, blockName, "metadata")
	entries, err := os.ReadDir(blockDir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
			return true
		}
	}
	return false
}

// storeMetadata stores block metadata to disk

func (pm *PackageManager) storeMetadata(metadata *BlockMetadata) error {
	// Store per-version at <block>/metadata/<version>.json
	metadataDir := filepath.Join(pm.InstallDir, metadata.Name, "metadata")
	if err := os.MkdirAll(metadataDir, 0755); err != nil {
		return fmt.Errorf("failed to create metadata directory: %w", err)
	}

	metadataPath := filepath.Join(metadataDir, fmt.Sprintf("%s.json", metadata.Version))
	file, err := os.Create(metadataPath)
	if err != nil {
		return fmt.Errorf("failed to create metadata file: %w", err)
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(metadata); err != nil {
		return fmt.Errorf("failed to encode metadata: %w", err)
	}

	return nil
}

// getMetadata retrieves block metadata from disk
func (pm *PackageManager) getMetadata(blockName string) (*BlockMetadata, error) {
	// Choose the most recently modified version metadata file
	blockDir := filepath.Join(pm.InstallDir, blockName, "metadata")
	entries, err := os.ReadDir(blockDir)
	if err != nil {
		return nil, fmt.Errorf("failed to open metadata directory: %w", err)
	}

	var latestPath string
	var latestMod int64
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		p := filepath.Join(blockDir, e.Name())
		info, err := os.Stat(p)
		if err != nil {
			continue
		}
		if info.ModTime().UnixNano() > latestMod {
			latestMod = info.ModTime().UnixNano()
			latestPath = p
		}
	}
	if latestPath == "" {
		return nil, fmt.Errorf("no metadata found for block %s", blockName)
	}

	file, err := os.Open(latestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open metadata file: %w", err)
	}
	defer file.Close()

	var metadata BlockMetadata
	if err := json.NewDecoder(file).Decode(&metadata); err != nil {
		return nil, fmt.Errorf("failed to decode metadata: %w", err)
	}

	return &metadata, nil
}

const (
	getDefaultInstallDirPathName = ".atomos"
)

// userHomeDir resolves the user's home directory reliably.
func userHomeDir() string {
	if homeDir, err := os.UserHomeDir(); err == nil && homeDir != "" {
		return homeDir
	}
	if envHome := os.Getenv("HOME"); envHome != "" {
		return envHome
	}
	// As an extreme fallback, use current working directory.
	if cwd, err := os.Getwd(); err == nil && cwd != "" {
		return cwd
	}

	return os.TempDir()
}

func getDefaultInstallDirPath() string {
	home := userHomeDir()
	return filepath.Join(home, getDefaultInstallDirPathName)
}

// loadExistingInstallation loads the existing installation state
func (pm *PackageManager) loadExistingInstallation() error {
	if !pm.isExistingInstallation() {
		return nil
	}

	if err := pm.checkBinariesExistAndLoad(); err != nil {
		return fmt.Errorf("installation failed: %w", err)
	}

	return nil
}

// checkBinariesExistAndLoad verifies that binaries referenced by installed blocks exist,
// and loads their metadata into memory if they do.
func (pm *PackageManager) checkBinariesExistAndLoad() error {
	listResult, err := pm.list()
	if err != nil {
		return fmt.Errorf("failed to list installed blocks: %w", err)
	}

	for _, block := range listResult.Blocks {
		if _, err := os.Stat(block.BinaryPath); os.IsNotExist(err) {
			return fmt.Errorf("block '%s' metadata exists but binary is missing: %s", block.Name, block.BinaryPath)
		}

		for _, block := range listResult.Blocks {
			pm.loadedBlocks[block.Name] = &block
		}
	}

	if len(listResult.Blocks) > 0 {
		fmt.Printf("Loaded existing AtomOS installation with %d blocks\n", len(listResult.Blocks))
	}

	return nil
}

// isExistingInstallation checks if this package manager is working with an existing installation
func (pm *PackageManager) isExistingInstallation() bool {
	if pm.loadedBlocks != nil {
		return len(pm.loadedBlocks) > 0
	}

	// Check if any block directory contains metadata files
	files, err := os.ReadDir(pm.InstallDir)
	if err != nil {
		return false
	}

	// If any block directory contains a versioned metadata file, it's an existing installation
	for _, file := range files {
		if file.IsDir() {
			blockDir := filepath.Join(pm.InstallDir, file.Name())
			metadataDir := filepath.Join(blockDir, "metadata")
			entries, _ := os.ReadDir(metadataDir)
			for _, e := range entries {
				if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
					return true
				}
			}
		}
	}

	return false
}

// list returns all installed blocks
func (pm *PackageManager) list() (*listResult, error) {
	// TODO: We likely don't want to do this on every call, make it a separate set up step instead.
	if err := os.MkdirAll(pm.InstallDir, 0755); err != nil {
		return nil, err
	}

	files, err := os.ReadDir(pm.InstallDir)
	if err != nil {
		return nil, err
	}

	var blocks []BlockMetadata
	for _, file := range files {
		if file.IsDir() {
			blockName := file.Name()
			metadata, err := pm.getMetadata(blockName)
			if err != nil {
				continue
			}
			blocks = append(blocks, *metadata)
		}
	}

	return &listResult{
		Blocks: blocks,
		Total:  len(blocks),
	}, nil
}
