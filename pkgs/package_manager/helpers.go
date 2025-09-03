package package_manager

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

	// Create install directory if it doesn't exist
	installDir := filepath.Join(pm.InstallDir, "binaries", blockInfo.Name)
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
