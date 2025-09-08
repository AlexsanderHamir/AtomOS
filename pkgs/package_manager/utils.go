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
	"time"
)

// getReleaseByTag fetches a specific GitHub release by tag and is tolerant
// to tags with or without a leading 'v'. Supports both public and private repos.
func (pm *PackageManager) getReleaseByTag(repo, tag string) (*GitHubRelease, error) {
	token := os.Getenv("GITHUB_TOKEN")
	client := &http.Client{Timeout: 30 * time.Second}

	withV := tag
	if !strings.HasPrefix(tag, "v") {
		withV = "v" + tag
	}
	withoutV := strings.TrimPrefix(tag, "v")

	for _, candidate := range []string{withV, withoutV} {
		url := fmt.Sprintf("https://api.github.com/repos/%s/releases/tags/%s", repo, candidate)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("create request for tag '%s': %w", candidate, err)
		}

		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		req.Header.Set("Accept", "application/vnd.github+json")

		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("fetch release by tag '%s': %w", candidate, err)
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("read response for tag '%s': %w", candidate, err)
		}

		switch resp.StatusCode {
		case http.StatusOK:
			var release GitHubRelease
			if err := json.Unmarshal(body, &release); err != nil {
				return nil, fmt.Errorf("decode JSON for tag '%s': %w", candidate, err)
			}
			return &release, nil

		case http.StatusNotFound:
			continue

		case http.StatusUnauthorized, http.StatusForbidden:
			return nil, fmt.Errorf("authentication failed for %s - check GITHUB_TOKEN", repo)

		default:
			return nil, fmt.Errorf("GitHub API error %d for tag '%s': %s",
				resp.StatusCode, candidate, strings.TrimSpace(string(body)))
		}
	}

	return nil, fmt.Errorf("release not found for tag '%s' in %s (tried with/without 'v')", tag, repo)
}

// getBinaryNameForPlatform returns the binary name for the current platform
func (pm *PackageManager) getBinaryNameForPlatform(blockInfo *BlockInfo) (string, error) {
	osName := runtime.GOOS
	arch := runtime.GOARCH
	platformKey := fmt.Sprintf("%s-%s", osName, arch)

	binaryName, exists := blockInfo.Binary.Assets[platformKey]
	if !exists {
		return "", fmt.Errorf("no binary found for platform %s", platformKey)
	}

	return binaryName, nil
}

// findAsset finds the asset by name and returns the asset object
func (pm *PackageManager) findAsset(release *GitHubRelease, assetName string) (*ReleaseAsset, error) {
	for _, asset := range release.Assets {
		if asset.Name == assetName {
			return &asset, nil
		}
	}
	return nil, fmt.Errorf("asset '%s' not found in release %s", assetName, release.TagName)
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
			Blockname := file.Name()
			metadata, err := pm.getMetadata(Blockname)
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
