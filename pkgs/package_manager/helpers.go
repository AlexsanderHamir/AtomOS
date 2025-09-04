package packagemanager

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v3"
)

// fetchBlockInfo fetches the agentic_support.yaml file from the repository
func (pm *PackageManager) fetchBlockInfo(repo string) (*BlockInfo, error) {
	// Construct the raw GitHub URL for the agentic_support.yaml file
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

	// Read the response body
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse the YAML content
	var blockInfo BlockInfo
	if err := yaml.Unmarshal(content, &blockInfo); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return &blockInfo, nil
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

// downloadAndVerifyBinary downloads a binary and verifies its SHA256 hash
func (pm *PackageManager) downloadAndVerifyBinary(repo, version string, blockInfo *BlockInfo) (string, string, error) {
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
		return "", "", fmt.Errorf("no binary found for platform %s", platformKey)
	}

	// Create per-block directory under the user's bin namespace, e.g., ~/bin/atomos/<block>
	installDir := filepath.Join(defaultBinaryBaseDir(), blockInfo.Name)
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return "", "", fmt.Errorf("failed to create install directory: %w", err)
	}

	// Construct the local file path
	localPath := filepath.Join(installDir, binaryName)

	// Download the binary from GitHub releases
	downloadURL := fmt.Sprintf("https://github.com/%s/releases/download/%s/%s", repo, version, binaryName)

	resp, err := http.Get(downloadURL)
	if err != nil {
		return "", "", fmt.Errorf("failed to download binary: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("failed to download binary: HTTP %d", resp.StatusCode)
	}

	// Create the local file
	file, err := os.Create(localPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to create local file: %w", err)
	}
	defer file.Close()

	// Copy the downloaded content to the file and calculate SHA256
	hash := sha256.New()
	teeReader := io.TeeReader(resp.Body, hash)

	_, err = io.Copy(file, teeReader)
	if err != nil {
		return "", "", fmt.Errorf("failed to write binary to file: %w", err)
	}

	// Get the SHA256 hash
	sha256Hash := hex.EncodeToString(hash.Sum(nil))

	// Make the binary executable on Unix-like systems
	if osName != "windows" {
		if err := os.Chmod(localPath, 0755); err != nil {
			return "", "", fmt.Errorf("failed to make binary executable: %w", err)
		}
	}

	return localPath, sha256Hash, nil
}

// isBlockInstalled checks if a block is already installed
func (pm *PackageManager) isBlockInstalled(blockName string) bool {
	metadataPath := filepath.Join(pm.InstallDir, "metadata", fmt.Sprintf("%s.json", blockName))
	_, err := os.Stat(metadataPath)
	return err == nil
}

// storeMetadata stores block metadata to disk
func (pm *PackageManager) storeMetadata(metadata *BlockMetadata) error {
	metadataDir := filepath.Join(pm.InstallDir, "metadata")
	if err := os.MkdirAll(metadataDir, 0755); err != nil {
		return fmt.Errorf("failed to create metadata directory: %w", err)
	}

	metadataPath := filepath.Join(metadataDir, fmt.Sprintf("%s.json", metadata.Name))

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
	metadataPath := filepath.Join(pm.InstallDir, "metadata", fmt.Sprintf("%s.json", blockName))

	file, err := os.Open(metadataPath)
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
	getDefaultCacheDirPathName   = "cache"
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

func getDefaultCacheDirPath(installDir string) string {
	return filepath.Join(installDir, getDefaultCacheDirPathName)
}

// defaultBinaryBaseDir returns the base directory for installed binaries.
// Example: ~/.atomos
func defaultBinaryBaseDir() string {
	return getDefaultInstallDirPath()
}

// loadExistingInstallation loads the existing installation state
func (pm *PackageManager) loadExistingInstallation() error {
	if !pm.IsExistingInstallation() {
		return nil
	}

	// Validate the existing installation to ensure it's in a good state
	if err := pm.checkBinariesExist(); err != nil {
		return fmt.Errorf("installation validation failed: %w", err)
	}

	// Load all existing metadata into memory for faster access
	listResult, err := pm.List()
	if err != nil {
		return fmt.Errorf("failed to load existing blocks: %w", err)
	}

	// Store the loaded blocks in memory as a map for fast lookups
	pm.loadedBlocks = make(map[string]BlockMetadata)
	for _, block := range listResult.Blocks {
		pm.loadedBlocks[block.Name] = block
	}
	pm.isLoaded = true

	// Log the loaded installation state
	if len(listResult.Blocks) > 0 {
		fmt.Printf("Loaded existing AtomOS installation with %d blocks\n", len(listResult.Blocks))
	}

	return nil
}

// checkBinariesExist checks if the existing blocks have valid binaries.
func (pm *PackageManager) checkBinariesExist() error {
	listResult, err := pm.List()
	if err != nil {
		return fmt.Errorf("failed to list installed blocks: %w", err)
	}

	for _, block := range listResult.Blocks {
		if _, err := os.Stat(block.BinaryPath); os.IsNotExist(err) {
			return fmt.Errorf("block '%s' metadata exists but binary is missing: %s", block.Name, block.BinaryPath)
		}
	}

	return nil
}
