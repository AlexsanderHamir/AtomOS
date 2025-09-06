package tests

import (
	"fmt"
	"os"
	"testing"

	packagemanager "github.com/AlexsanderHamir/AtomOS/pkgs/package_manager"
)

func TestInstallWithTestDir(t *testing.T) {
	t.Parallel()
	testDir := fmt.Sprintf("./atomos-test-dir-%s", t.Name())
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test dir: %s", err)
	}
	defer os.RemoveAll(testDir)

	pkgm := packagemanager.NewPackageManagerWithTestDir(testDir)
	if pkgm == nil {
		t.Fatal("package manager can't be nil")
	}

	var blockMetaData *packagemanager.BlockMetadata

	t.Run("InstallBlock", func(t *testing.T) {
		var err error
		installReq := packagemanager.InstallRequest{Repo: "AlexsanderHamir/prof", Version: "1.8.1"}
		blockMetaData, err = pkgm.Install(installReq)
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
	})

	t.Run("GetBlock", func(t *testing.T) {
		newBlockMetadata, ok := pkgm.GetLoadedBlock(blockMetaData.Name)
		if !ok {
			t.Fatal("block should be present")
		}

		if newBlockMetadata == nil {
			t.Fatal("block metadata can't be nil")
		}

		CompareBlockMetadata(t, blockMetaData, newBlockMetadata)

		t.Run("GetUnknownBlock", func(t *testing.T) {
			newBlockMetadata, ok := pkgm.GetLoadedBlock("fake_block")
			if ok {
				t.Fatal("block shouldn't be present")
			}

			if newBlockMetadata != nil {
				t.Fatal("block metadata should be nil")
			}
		})
	})


	
}

func TestInstallVersionWithoutAgenticSupport(t *testing.T) {
	t.Parallel()
	testDir := fmt.Sprintf("./atomos-test-dir-%s", t.Name())
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test dir: %s", err)
	}
	defer os.RemoveAll(testDir)

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
}
