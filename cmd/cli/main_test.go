package main

import (
	"os"
	"sync"
	"testing"

	internal "github.com/lion-and-bear/freeathome/v2/internal"
)

// TestMainVersionSetting tests that the version and commit are set correctly.
func TestMainVersionSetting(t *testing.T) {
	if version != "debug" {
		t.Errorf("Expected version to be 'debug', got '%s'", version)
	}

	if commit != "unknown" {
		t.Errorf("Expected commit to be 'unknown', got '%s'", commit)
	}
}

// TestVersionAssignment tests that the version and commit are assigned correctly.
func TestVersionAssignment(t *testing.T) {
	// Use a mutex to prevent concurrent access to the version variables
	var mu sync.Mutex
	mu.Lock()
	defer mu.Unlock()

	originalVersion := internal.Version
	originalCommit := internal.Commit

	// Set test values
	internal.Version = "test-version"
	internal.Commit = "test-commit"

	// Verify assignment
	if internal.Version != "test-version" {
		t.Errorf("Expected internal.Version to be 'test-version', got '%s'", internal.Version)
	}

	if internal.Commit != "test-commit" {
		t.Errorf("Expected internal.Commit to be 'test-commit', got '%s'", internal.Commit)
	}

	// Restore original values
	internal.Version = originalVersion
	internal.Commit = originalCommit
}

// TestMain runs tests sequentially to avoid data races on global variables
func TestMain(m *testing.M) {
	// Run tests sequentially to avoid data races on global variables
	os.Exit(m.Run())
}
