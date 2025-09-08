// Copyright (c) 2025 Alexsander Hamir Gomes Baptista
//
// This file is part of AtomOS and licensed under the Sustainable Use License (SUL).
// You may use, modify, and redistribute this software for personal or internal business use.
// Offering it as a commercial hosted service requires a separate license.
//
// Full license: see the LICENSE file in the root of this repository
// or contact alexsanderhamirgomesbaptista@gmail.com.

package main

import (
	"fmt"
	"log"

	packagemanager "github.com/AlexsanderHamir/AtomOS/pkgs/package_manager"
)

func main() {
	pm := packagemanager.NewPackageManager()
	
	// Test installing the block to see what LSP entries are captured
	installReq := packagemanager.InstallRequest{Repo: "AlexsanderHamir/prof", Version: "1.8.1", Force: true}
	blockMetadata, err := pm.Install(installReq)
	if err != nil {
		log.Fatalf("Failed to install: %v", err)
	}
	
	fmt.Printf("BlockMetadata: %+v\n", blockMetadata)
	fmt.Printf("LSP Entries count: %d\n", len(blockMetadata.LSPEntries))
	
	for name, entry := range blockMetadata.LSPEntries {
		fmt.Printf("Entry '%s': %+v\n", name, entry)
	}
}
