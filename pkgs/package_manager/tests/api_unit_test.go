package tests

import (
	"os"
	"path/filepath"
	"testing"

	packagemanager "github.com/AlexsanderHamir/AtomOS/pkgs/package_manager"
)

func TestInstallWithTestDir(t *testing.T) {
	testDir := "./atomos-test-dir"
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test dir: %s", err)
	}
	defer os.RemoveAll(testDir)

	pkgm := packagemanager.NewPackageManagerWithTestDir(testDir)
	if pkgm == nil {
		t.Fatal("package manager can't be nil")
	}

	installReq := packagemanager.InstallRequest{Repo: "AlexsanderHamir/prof", Version: "1.8.1"}
	blockMetaData, err := pkgm.Install(installReq)
	if err != nil {
		t.Fatalf("pkgm.Install() failed: %s", err)
	}
	if blockMetaData == nil {
		t.Fatal("block metadata can't be nil")
	}

	verifyDirectoryStructure(t, testDir)
	verifyMetadataFile(t, testDir, blockMetaData)
	verifyBinaryExecution(t, blockMetaData)
	verifyLSPEntries(t, blockMetaData)
}

func TestInstallVersionWithoutAgenticSupport(t *testing.T) {
	testDir := "./atomos-test-dir-v180"
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test dir: %s", err)
	}
	defer os.RemoveAll(testDir) // Clean up after test

	pkgm := packagemanager.NewPackageManagerWithTestDir(testDir)
	if pkgm == nil {
		t.Fatal("package manager can't be nil")
	}

	installReq := packagemanager.InstallRequest{Repo: "AlexsanderHamir/prof", Version: "1.8.0"}
	blockMetaData, err := pkgm.Install(installReq)
	if err == nil {
		t.Fatal("Expected installation to fail for version 1.8.0 (no agentic_support.yaml), but it succeeded")
	}

	if blockMetaData != nil {
		t.Fatal("Expected block metadata to be nil when installation fails")
	}

	// Verify that no files were created in the test directory
	if _, err := os.Stat(filepath.Join(testDir, ".atomos")); !os.IsNotExist(err) {
		t.Fatal("Expected no .atomos directory to be created when installation fails")
	}
}
