package tests

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	packagemanager "github.com/AlexsanderHamir/AtomOS/pkgs/package_manager"
)

const expectedLSPEntriesNum = 3

var expectedLSPEntries = map[string]packagemanager.Entry{
	"run": {
		Name:        "run",
		Command:     "prof run",
		Description: "Run profiling on the target binary",
		Inputs: []packagemanager.Input{
			{Name: "target", Type: "path"},
		},
		Outputs: []packagemanager.Output{
			{Name: "profile", Type: "file"},
		},
	},
	"report": {
		Name:        "report",
		Command:     "prof report",
		Description: "Generate a profiling report from a saved profile",
		Inputs: []packagemanager.Input{
			{Name: "profile", Type: "file"},
		},
		Outputs: []packagemanager.Output{
			{Name: "summary", Type: "string"},
		},
	},
	"flamegraph": {
		Name:        "flamegraph",
		Command:     "prof flamegraph",
		Description: "Generate a flamegraph from a profile file",
		Inputs: []packagemanager.Input{
			{Name: "profile", Type: "file"},
		},
		Outputs: []packagemanager.Output{
			{Name: "flamegraph", Type: "svg"},
		},
	},
}

func verifyDirectoryStructure(t *testing.T, testDir string) {
	goProfilerDir := filepath.Join(testDir, ".atomos", "go-profiler")

	// Check that the go-profiler directory exists
	if _, err := os.Stat(goProfilerDir); os.IsNotExist(err) {
		t.Fatalf("go-profiler directory does not exist: %s", goProfilerDir)
	}

	// Check that the bin directory exists
	binDir := filepath.Join(goProfilerDir, "bin")
	if _, err := os.Stat(binDir); os.IsNotExist(err) {
		t.Fatalf("bin directory does not exist: %s", binDir)
	}

	// Check that the metadata directory exists
	metadataDir := filepath.Join(goProfilerDir, "metadata")
	if _, err := os.Stat(metadataDir); os.IsNotExist(err) {
		t.Fatalf("metadata directory does not exist: %s", metadataDir)
	}

	// Check that there's at least one binary file in the bin directory
	binEntries, err := os.ReadDir(binDir)
	if err != nil {
		t.Fatalf("Failed to read bin directory: %s", err)
	}
	if len(binEntries) == 0 {
		t.Fatal("bin directory is empty, expected at least one binary file")
	}
}

func verifyMetadataFile(t *testing.T, testDir string, blockMetaData *packagemanager.BlockMetadata) {
	metadataFile := filepath.Join(testDir, ".atomos", "go-profiler", "metadata", "1.8.1.json")

	if _, err := os.Stat(metadataFile); os.IsNotExist(err) {
		t.Fatalf("metadata file for version 1.8.1 does not exist: %s", metadataFile)
	}

	fileMetadata := readAndDecodeMetadata(t, metadataFile)

	compareMetadataFields(t, blockMetaData, fileMetadata)

	verifyExpectedValues(t, fileMetadata, testDir)
}

func readAndDecodeMetadata(t *testing.T, metadataFile string) *packagemanager.BlockMetadata {
	metadataContent, err := os.ReadFile(metadataFile)
	if err != nil {
		t.Fatalf("Failed to read metadata file: %s", err)
	}
	if len(metadataContent) == 0 {
		t.Fatal("metadata file is empty")
	}

	var fileMetadata packagemanager.BlockMetadata
	if err := json.Unmarshal(metadataContent, &fileMetadata); err != nil {
		t.Fatalf("Failed to parse metadata JSON: %s", err)
	}

	return &fileMetadata
}

func compareMetadataFields(t *testing.T, returned, file *packagemanager.BlockMetadata) {
	if returned.Name != file.Name {
		t.Fatalf("Name mismatch: returned='%s', file='%s'", returned.Name, file.Name)
	}

	if returned.Version != file.Version {
		t.Fatalf("Version mismatch: returned='%s', file='%s'", returned.Version, file.Version)
	}

	if returned.SourceRepo != file.SourceRepo {
		t.Fatalf("SourceRepo mismatch: returned='%s', file='%s'", returned.SourceRepo, file.SourceRepo)
	}

	if returned.BinaryPath != file.BinaryPath {
		t.Fatalf("BinaryPath mismatch: returned='%s', file='%s'", returned.BinaryPath, file.BinaryPath)
	}
}

func verifyExpectedValues(t *testing.T, metadata *packagemanager.BlockMetadata, testDir string) {
	expectedName := "go-profiler"
	if metadata.Name != expectedName {
		t.Fatalf("Expected name to be '%s', got '%s'", expectedName, metadata.Name)
	}

	expectedVersion := "1.8.1"
	if metadata.Version != expectedVersion {
		t.Fatalf("Expected version to be '%s', got '%s'", expectedVersion, metadata.Version)
	}

	expectedSourceRepo := "AlexsanderHamir/prof"
	if metadata.SourceRepo != expectedSourceRepo {
		t.Fatalf("Expected source_repo to be '%s', got '%s'", expectedSourceRepo, metadata.SourceRepo)
	}

	// Check binary_path points to the expected directory structure
	expectedPathPrefix := filepath.Join(testDir, ".atomos", "go-profiler", "bin")
	normalizedBinaryPath := filepath.Clean(metadata.BinaryPath)
	normalizedExpectedPrefix := filepath.Clean(expectedPathPrefix)
	if !strings.HasPrefix(normalizedBinaryPath, normalizedExpectedPrefix) {
		t.Fatalf("Expected binary_path to start with '%s', got '%s'", normalizedExpectedPrefix, normalizedBinaryPath)
	}
}

func verifyBinaryExecution(t *testing.T, blockMetaData *packagemanager.BlockMetadata) {
	cmd := exec.Command(blockMetaData.BinaryPath, "--help")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to execute binary '%s': %s, output: %s", blockMetaData.BinaryPath, err, string(output))
	}

	if !strings.Contains(string(output), "CLI tool for organizing pprof generated data, and analyzing performance differences at the profile level.") {
		t.Fatal("Binary execution produced no output")
	}
}

func verifyLSPEntryDetails(t *testing.T, entryName string, entry, expectedEntry packagemanager.Entry) {
	if entry.Name != expectedEntry.Name {
		t.Fatalf("LSP entry '%s' name mismatch: expected='%s', got='%s'", entryName, expectedEntry.Name, entry.Name)
	}
	if entry.Command != expectedEntry.Command {
		t.Fatalf("LSP entry '%s' command mismatch: expected='%s', got='%s'", entryName, expectedEntry.Command, entry.Command)
	}
	if entry.Description != expectedEntry.Description {
		t.Fatalf("LSP entry '%s' description mismatch: expected='%s', got='%s'", entryName, expectedEntry.Description, entry.Description)
	}
}

func verifyLSPEntryInputs(t *testing.T, entryName string, entry, expectedEntry packagemanager.Entry) {
	if len(entry.Inputs) != len(expectedEntry.Inputs) {
		t.Fatalf("LSP entry '%s' inputs count mismatch: expected=%d, got=%d", entryName, len(expectedEntry.Inputs), len(entry.Inputs))
	}
	for i, expectedInput := range expectedEntry.Inputs {
		if i >= len(entry.Inputs) {
			t.Fatalf("LSP entry '%s' missing input at index %d", entryName, i)
		}
		if entry.Inputs[i].Name != expectedInput.Name {
			t.Fatalf("LSP entry '%s' input %d name mismatch: expected='%s', got='%s'", entryName, i, expectedInput.Name, entry.Inputs[i].Name)
		}
		if entry.Inputs[i].Type != expectedInput.Type {
			t.Fatalf("LSP entry '%s' input %d type mismatch: expected='%s', got='%s'", entryName, i, expectedInput.Type, entry.Inputs[i].Type)
		}
	}
}

func verifyLSPEntryOutputs(t *testing.T, entryName string, entry, expectedEntry packagemanager.Entry) {
	if len(entry.Outputs) != len(expectedEntry.Outputs) {
		t.Fatalf("LSP entry '%s' outputs count mismatch: expected=%d, got=%d", entryName, len(expectedEntry.Outputs), len(entry.Outputs))
	}
	for i, expectedOutput := range expectedEntry.Outputs {
		if i >= len(entry.Outputs) {
			t.Fatalf("LSP entry '%s' missing output at index %d", entryName, i)
		}
		if entry.Outputs[i].Name != expectedOutput.Name {
			t.Fatalf("LSP entry '%s' output %d name mismatch: expected='%s', got='%s'", entryName, i, expectedOutput.Name, entry.Outputs[i].Name)
		}
		if entry.Outputs[i].Type != expectedOutput.Type {
			t.Fatalf("LSP entry '%s' output %d type mismatch: expected='%s', got='%s'", entryName, i, expectedOutput.Type, entry.Outputs[i].Type)
		}
	}
}

func verifyLSPEntries(t *testing.T, blockMetaData *packagemanager.BlockMetadata) {
	if len(blockMetaData.LSPEntries) < expectedLSPEntriesNum {
		t.Fatal("missing lsp entries")
	}

	for entryName, expectedEntry := range expectedLSPEntries {
		entry, exists := blockMetaData.LSPEntries[entryName]
		if !exists {
			t.Fatalf("Expected LSP entry '%s' not found", entryName)
		}

		verifyLSPEntryDetails(t, entryName, entry, expectedEntry)
		verifyLSPEntryInputs(t, entryName, entry, expectedEntry)
		verifyLSPEntryOutputs(t, entryName, entry, expectedEntry)
	}

	if len(blockMetaData.LSPEntries) != len(expectedLSPEntries) {
		t.Fatalf("LSP entries count mismatch: expected=%d, got=%d", len(expectedLSPEntries), len(blockMetaData.LSPEntries))
	}
}

// CompareBlockMetadata compares two BlockMetadata instances for equality
func CompareBlockMetadata(t *testing.T, original, retrieved *packagemanager.BlockMetadata) {
	if original == nil || retrieved == nil {
		t.Fatal("Both block metadata instances must be non-nil")
	}

	if original.Name != retrieved.Name {
		t.Errorf("Name mismatch: expected %s, got %s", original.Name, retrieved.Name)
	}

	if original.Version != retrieved.Version {
		t.Errorf("Version mismatch: expected %s, got %s", original.Version, retrieved.Version)
	}

	if original.SourceRepo != retrieved.SourceRepo {
		t.Errorf("SourceRepo mismatch: expected %s, got %s", original.SourceRepo, retrieved.SourceRepo)
	}

	if original.BinaryPath != retrieved.BinaryPath {
		t.Errorf("BinaryPath mismatch: expected %s, got %s", original.BinaryPath, retrieved.BinaryPath)
	}

	if original.IsActive != retrieved.IsActive {
		t.Errorf("IsActive mismatch: expected %t, got %t", original.IsActive, retrieved.IsActive)
	}

	// Compare timestamps using the existing helper function
	CompareBlockTimestamps(t, original, retrieved)

	// Compare LSP entries
	if len(original.LSPEntries) != len(retrieved.LSPEntries) {
		t.Errorf("LSPEntries count mismatch: expected %d, got %d", len(original.LSPEntries), len(retrieved.LSPEntries))
	}

	for name, originalEntry := range original.LSPEntries {
		retrievedEntry, exists := retrieved.LSPEntries[name]
		if !exists {
			t.Errorf("LSP entry %s missing in retrieved metadata", name)
			continue
		}

		if originalEntry.Name != retrievedEntry.Name {
			t.Errorf("LSP entry %s name mismatch: expected %s, got %s", name, originalEntry.Name, retrievedEntry.Name)
		}

		if originalEntry.Command != retrievedEntry.Command {
			t.Errorf("LSP entry %s command mismatch: expected %s, got %s", name, originalEntry.Command, retrievedEntry.Command)
		}

		if originalEntry.Description != retrievedEntry.Description {
			t.Errorf("LSP entry %s description mismatch: expected %s, got %s", name, originalEntry.Description, retrievedEntry.Description)
		}
	}
}

// CompareBlockTimestamps compares the timestamp fields of two BlockMetadata instances
func CompareBlockTimestamps(t *testing.T, original, retrieved *packagemanager.BlockMetadata) {
	if original == nil || retrieved == nil {
		t.Fatal("Both block metadata instances must be non-nil")
	}

	// InstalledAt should be the same (or very close)
	timeDiff := original.InstalledAt.Sub(retrieved.InstalledAt)
	if timeDiff < 0 {
		timeDiff = -timeDiff
	}
	if timeDiff > time.Second {
		t.Errorf("InstalledAt time difference too large: %v", timeDiff)
	}

	// LastUpdated should be the same (or very close)
	timeDiff = original.LastUpdated.Sub(retrieved.LastUpdated)
	if timeDiff < 0 {
		timeDiff = -timeDiff
	}
	if timeDiff > time.Second {
		t.Errorf("LastUpdated time difference too large: %v", timeDiff)
	}
}
