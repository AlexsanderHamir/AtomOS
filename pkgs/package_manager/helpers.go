package packagemanager

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type GitHubAsset struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	DownloadURL string `json:"browser_download_url"`
}

type githubContent struct {
	Content  string `json:"content"`
	Encoding string `json:"encoding"`
}

func (pm *PackageManager) fetchBlockInfo(repo string) (*BlockInfo, error) {
	token := os.Getenv("GITHUB_TOKEN")
	client := &http.Client{}

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/contents/agentic_support.yaml", repo)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch agentic_support.yaml: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		switch resp.StatusCode {
		case http.StatusNotFound:
			return nil, fmt.Errorf("agentic_support.yaml not found in repository %s", repo)
		case http.StatusUnauthorized, http.StatusForbidden:
			return nil, fmt.Errorf("authentication failed - check GITHUB_TOKEN permissions for repository %s", repo)
		default:
			return nil, fmt.Errorf("GitHub API error %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
		}
	}

	var gc githubContent
	if err := json.Unmarshal(body, &gc); err != nil {
		return nil, fmt.Errorf("failed to parse GitHub API response: %w", err)
	}

	if gc.Encoding != "base64" {
		return nil, fmt.Errorf("unexpected encoding: %s", gc.Encoding)
	}

	data, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(gc.Content, "\n", ""))
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 content: %w", err)
	}

	var blockInfo BlockInfo
	if err := yaml.Unmarshal(data, &blockInfo); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return &blockInfo, nil
}

// getLatestRelease fetches the latest release from GitHub (supports both public and private repos)
func (pm *PackageManager) getLatestRelease(repo string) (*GitHubRelease, error) {
	token := os.Getenv("GITHUB_TOKEN")
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch latest release: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		switch resp.StatusCode {
		case http.StatusNotFound:
			return nil, fmt.Errorf("no releases found for repository %s", repo)
		case http.StatusUnauthorized, http.StatusForbidden:
			return nil, fmt.Errorf("authentication failed - check GITHUB_TOKEN permissions for repository %s", repo)
		default:
			return nil, fmt.Errorf("GitHub API error %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
		}
	}

	var release GitHubRelease
	if err := json.Unmarshal(body, &release); err != nil {
		return nil, fmt.Errorf("failed to decode release JSON: %w", err)
	}

	return &release, nil
}

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

// downloadBinary downloads a binary for the current platform
func (pm *PackageManager) downloadBinary(repo, version string, blockInfo *BlockInfo) (string, error) {
	binaryName, err := pm.getBinaryNameForPlatform(blockInfo)
	if err != nil {
		return "", err
	}

	binDir := filepath.Join(pm.InstallDir, blockInfo.Name, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create bin directory: %w", err)
	}

	localPath := filepath.Join(binDir, binaryName)

	if err := pm.downloadAsset(repo, version, binaryName, localPath); err != nil {
		return "", fmt.Errorf("downloadAsset failed: %w", err)
	}

	if runtime.GOOS != "windows" {
		if err := os.Chmod(localPath, 0755); err != nil {
			return "", fmt.Errorf("failed to make binary executable: %w", err)
		}
	}

	return localPath, nil
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

// downloadAsset downloads a specific asset from a GitHub release
func (pm *PackageManager) downloadAsset(repo, version, assetName, localPath string) error {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return fmt.Errorf("GITHUB_TOKEN is required for downloading assets")
	}

	// Get release to find asset
	release, err := pm.getReleaseByTag(repo, version)
	if err != nil {
		return fmt.Errorf("failed to resolve release '%s': %w", version, err)
	}

	// Find the asset (not just the URL).
	asset, err := pm.findAsset(release, assetName)
	if err != nil {
		return fmt.Errorf("findAsset failed: %w", err)
	}

	// Use the GitHub API endpoint with asset ID.
	assetURL := fmt.Sprintf("https://api.github.com/repos/%s/releases/assets/%d", repo, asset.ID)

	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("GET", assetURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create asset request: %w", err)
	}

	// Required headers for GitHub asset downloads
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/octet-stream") // Critical for binary downloads

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download asset: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("download failed: HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	// Create the local file
	file, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create local file: %w", err)
	}
	defer file.Close()

	// Copy the downloaded content to the file
	if _, err := io.Copy(file, resp.Body); err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	return nil
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

// isBlockInstalled checks if a block is already installed
func (pm *PackageManager) isBlockInstalled(Blockname string) bool {
	// Consider installed if there's at least one versioned metadata file under <block>/metadata/
	blockDir := filepath.Join(pm.InstallDir, Blockname, "metadata")
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
func (pm *PackageManager) getMetadata(Blockname string) (*BlockMetadata, error) {
	// Choose the most recently modified version metadata file
	blockDir := filepath.Join(pm.InstallDir, Blockname, "metadata")
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
		return nil, fmt.Errorf("no metadata found for block %s", Blockname)
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

// convertEntriesToMap converts a slice of Entry to a map[string]Entry using the entry name as the key
func convertEntriesToMap(entries []Entry) map[string]Entry {
	result := make(map[string]Entry)
	for _, entry := range entries {
		result[entry.Name] = entry
	}
	return result
}
