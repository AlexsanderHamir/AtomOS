package blocks

import (
	"strings"
	"testing"
)

func TestFetchBlockFromRepository(t *testing.T) {
	// Test fetching from a repository
	// Using the same repository that's referenced in the local agentic_support.yaml
	repo := "AlexsanderHamir/prof"

	block, err := FetchBlock(repo)
	if err != nil {
		// If the repository doesn't have an agentic_support.yaml file, skip the test
		if strings.Contains(err.Error(), "HTTP 404") {
			t.Skipf("Repository %s does not have an agentic_support.yaml file, skipping test", repo)
		}
		t.Fatalf("FetchBlock failed with error: %v", err)
	}

	if block == nil {
		t.Fatal("FetchBlock returned nil block")
	}

	// Verify that we got a valid block structure
	if block.Name == "" {
		t.Fatal("Block name is empty")
	}

	if block.Version == "" {
		t.Fatal("Block version is empty")
	}

	if block.Source.Repo == "" {
		t.Fatal("Block source repo is empty")
	}

	t.Logf("Successfully fetched block from repository: %s (version: %s)", block.Name, block.Version)
}
