package cmd

import (
	"bytes"
	"os"
	"slices"
	"strings"
	"testing"

	internal "github.com/lion-and-bear/freeathome/v2/internal"
	"github.com/spf13/cobra"
)

// TestVersionCommand tests that the version command has the expected properties.
func TestVersionCommand(t *testing.T) {
	// Test that version command has the expected properties
	if versionCmd.Use != "version" {
		t.Errorf("Expected version command Use to be 'version', got '%s'", versionCmd.Use)
	}

	if versionCmd.Short == "" {
		t.Error("Expected version command to have a Short description")
	}

	if versionCmd.Long == "" {
		t.Error("Expected version command to have a Long description")
	}
}

// TestVersionCommandOutput tests that the version command outputs the expected version information.
func TestVersionCommandOutput(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run the version command
	versionCmd.Run(versionCmd, []string{})

	// Restore stdout
	_ = w.Close()
	os.Stdout = oldStdout

	// Read the output
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// Check that output contains expected version information
	if !strings.Contains(output, "free@home CLI v") {
		t.Errorf("Expected output to contain 'free@home CLI v', got: %s", output)
	}

	if !strings.Contains(output, internal.Version) {
		t.Errorf("Expected output to contain version '%s', got: %s", internal.Version, output)
	}

	if !strings.Contains(output, internal.Commit) {
		t.Errorf("Expected output to contain commit '%s', got: %s", internal.Commit, output)
	}
}

// TestVersionCommandIsChildOfRoot tests that the version command is properly added to the root command.
func TestVersionCommandIsChildOfRoot(t *testing.T) {
	found := slices.ContainsFunc(rootCmd.Commands(), func(cmd *cobra.Command) bool {
		return cmd.Name() == "version"
	})
	if !found {
		t.Error("Expected version command to be a child of root command")
	}
}
