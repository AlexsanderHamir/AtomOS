// Copyright (c) 2025 Alexsander Hamir Gomes Baptista
//
// This file is part of AtomOS and licensed under the Sustainable Use License (SUL).
// You may use, modify, and redistribute this software for personal or internal business use.
// Offering it as a commercial hosted service requires a separate license.
//
// Full license: see the LICENSE file in the root of this repository
// or contact alexsanderhamirgomesbaptista@gmail.com.

package tests

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	packagemanager "github.com/AlexsanderHamir/AtomOS/pkgs/package_manager"
	"github.com/joho/godotenv"
)

func TestMain(m *testing.M) {
	err := godotenv.Load("../../../.env")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load .env: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()
	os.Exit(code)
}

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

	t.Run("InstallSupportedBlock", func(t *testing.T) {
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

		t.Run("InstallNonSupportedBlock", func(t *testing.T) {
			installReq := packagemanager.InstallRequest{Repo: "AlexsanderHamir/prof", Version: "1.8.0", Force: true}
			blockMetaData, err := pkgm.Install(installReq)
			if err == nil {
				t.Fatal("Expected installation to fail for version 1.8.0 (no agentic_support.yaml), but it succeeded")
			}

			if blockMetaData != nil {
				t.Fatal("Expected block metadata to be nil when installation fails")
			}
		})
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

	t.Run("Uninstall", func(t *testing.T) {
		err := pkgm.Uninstall(blockMetaData.Name)
		if err != nil {
			t.Fatalf("pkgm.Uninstall() failed: %s", err)
		}

		if _, err := os.Stat(blockMetaData.BinaryPath); !os.IsNotExist(err) {
			t.Fatalf("Binary file should be removed: %s", blockMetaData.BinaryPath)
		}

		metadataPath := filepath.Join(testDir, ".atomos", blockMetaData.Name, "metadata", fmt.Sprintf("%s.json", blockMetaData.Version))
		if _, err := os.Stat(metadataPath); !os.IsNotExist(err) {
			t.Fatalf("Metadata file should be removed: %s", metadataPath)
		}

		_, ok := pkgm.GetLoadedBlock(blockMetaData.Name)
		if ok {
			t.Fatal("Block should be removed from loaded blocks")
		}

		t.Run("UninstallNonExistentBlock", func(t *testing.T) {
			err := pkgm.Uninstall("non-existent-block")
			if err == nil {
				t.Fatal("Expected error when uninstalling non-existent block")
			}
			if !strings.Contains(err.Error(), "is not installed") {
				t.Fatalf("Expected error message to contain 'is not installed', got: %s", err.Error())
			}
		})
	})

}
