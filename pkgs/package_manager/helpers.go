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

// isBlockInstalled checks if there's at least one versioned metadata file under <block>/metadata/
func (pm *PackageManager) isBlockInstalled(Blockname string) bool {
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

// convertEntriesToMap converts a slice of Entry to a map[string]Entry using the entry name as the key
func convertEntriesToMap(entries []Entry) map[string]Entry {
	result := make(map[string]Entry)
	for _, entry := range entries {
		result[entry.Name] = entry
	}
	return result
}
