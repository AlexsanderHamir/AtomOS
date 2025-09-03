package blockmanager

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"gopkg.in/yaml.v3"
)

// fetchAgenticSupportYAML fetches the agentic_support.yaml file from the root of the provided repository
func fetchAgenticSupportYAML(repo string) ([]byte, error) {
	// Construct the raw GitHub URL for the agentic_support.yaml file
	// Assuming the repo is in format "owner/repo"
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

	return content, nil
}

func getBlockFromRepo(filePath string) (*BlockInfo, error) {
	var data []byte
	var err error

	// Check if filePath is a repository (contains slash) or a local file
	if strings.Contains(filePath, "/") && !strings.HasPrefix(filePath, "http") {
		// Treat as repository and fetch the YAML file
		data, err = fetchAgenticSupportYAML(filePath)
		if err != nil {
			return nil, err
		}
	} else {
		// Treat as local file path
		data, err = os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}
	}

	var block BlockInfo
	if err := yaml.Unmarshal(data, &block); err != nil {
		return nil, err
	}

	return &block, nil
}

// downloadAndStoreBinary downloads the appropriate binary for the current platform
// and returns the local file path where it was stored
func downloadAndStoreBinary(block *BlockInfo) (string, error) {
	osName := runtime.GOOS
	arch := runtime.GOARCH

	// Map Go's runtime values to the YAML keys
	platformKey := fmt.Sprintf("%s-%s", osName, arch)

	// Handle Windows executable extension
	if osName == "windows" {
		platformKey = fmt.Sprintf("%s-%s", osName, arch)
	}

	binaryName, exists := block.Binary.Assets[platformKey]
	if !exists {
		return "", fmt.Errorf("no binary found for platform %s", platformKey)
	}

	// Create downloads directory if it doesn't exist
	downloadDir := filepath.Join("downloads", block.Name)
	if err := os.MkdirAll(downloadDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create download directory: %w", err)
	}

	// Construct the local file path
	localPath := filepath.Join(downloadDir, binaryName)

	// Check if binary already exists
	if _, err := os.Stat(localPath); err == nil {
		return localPath, nil
	}

	// Download the binary from GitHub releases
	downloadURL := fmt.Sprintf("https://github.com/%s/releases/download/%s/%s",
		block.Source.Repo, block.Version, binaryName)

	resp, err := http.Get(downloadURL)
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
