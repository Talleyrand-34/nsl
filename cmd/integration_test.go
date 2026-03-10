/*
Copyright © 2025 Tecdesoft (rodrigo-gonzalez@tecdesoft.es, t34@t34.dev)

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published
by the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.
*/
package cmd_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCLI_HelpCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping CLI integration test in short mode")
	}

	// Build the CLI binary for testing
	binaryPath := buildCLIBinary(t)
	defer os.Remove(binaryPath)

	// Test basic help command
	cmd := exec.Command(binaryPath, "--help")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()

	assert.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, "Usage:")
	assert.Contains(t, output, "Available Commands:")
	assert.Contains(t, output, "diagram")
	assert.Contains(t, output, "modify")
	assert.Contains(t, output, "print")
	assert.Contains(t, output, "server")
}

func TestCLI_ServerHelpCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping CLI integration test in short mode")
	}

	binaryPath := buildCLIBinary(t)
	defer os.Remove(binaryPath)

	// Test server subcommand help
	cmd := exec.Command(binaryPath, "server", "--help")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()

	assert.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, "Print info about any table in the db")
	assert.Contains(t, output, "--port")
}

func TestCLI_DiagramHelpCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping CLI integration test in short mode")
	}

	binaryPath := buildCLIBinary(t)
	defer os.Remove(binaryPath)

	// Test diagram subcommand help
	cmd := exec.Command(binaryPath, "diagram", "--help")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()

	assert.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, "diagram")
}

func TestCLI_PrintHelpCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping CLI integration test in short mode")
	}

	binaryPath := buildCLIBinary(t)
	defer os.Remove(binaryPath)

	// Test print subcommand help
	cmd := exec.Command(binaryPath, "print", "--help")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()

	assert.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, "print")
	assert.Contains(t, output, "Print info about any table in the db")
}

func TestCLI_ModifyHelpCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping CLI integration test in short mode")
	}

	binaryPath := buildCLIBinary(t)
	defer os.Remove(binaryPath)

	// Test modify subcommand help
	cmd := exec.Command(binaryPath, "modify", "--help")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()

	assert.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, "modify")
	assert.Contains(t, output, "Available Commands:")
}

func TestCLI_InvalidCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping CLI integration test in short mode")
	}

	binaryPath := buildCLIBinary(t)
	defer os.Remove(binaryPath)

	// Test invalid command
	cmd := exec.Command(binaryPath, "invalid-command")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()

	// Should return error for invalid command
	assert.Error(t, err)
}

func TestCLI_VersionFlag(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping CLI integration test in short mode")
	}

	binaryPath := buildCLIBinary(t)
	defer os.Remove(binaryPath)

	// Test version information (using verbose flag as proxy since no explicit version flag)
	cmd := exec.Command(binaryPath, "--verbose", "--help")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()

	assert.NoError(t, err)

	output := out.String()
	assert.Contains(t, output, "nsl-graph")
}

// Helper function to build the CLI binary for testing
func buildCLIBinary(t *testing.T) string {
	// Get the project root directory
	wd, err := os.Getwd()
	assert.NoError(t, err)

	// Go up to project root (assuming we're in cmd/ directory)
	projectRoot := filepath.Dir(wd)

	// Create temporary binary
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "nsl-graph-test")

	// Build the binary
	cmd := exec.Command("go", "build", "-o", binaryPath, filepath.Join(projectRoot, "main.go"))
	cmd.Dir = projectRoot

	var buildOutput bytes.Buffer
	cmd.Stdout = &buildOutput
	cmd.Stderr = &buildOutput

	err = cmd.Run()
	if err != nil {
		t.Fatalf("Failed to build CLI binary: %v\nOutput: %s", err, buildOutput.String())
	}

	return binaryPath
}

// Test CLI with custom database path
func TestCLI_CustomDatabasePath(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping CLI integration test in short mode")
	}

	binaryPath := buildCLIBinary(t)
	defer os.Remove(binaryPath)

	// Create temporary database file
	tmpDB := filepath.Join(t.TempDir(), "test.db")

	// Test with custom database path
	cmd := exec.Command(binaryPath, "--source", tmpDB, "print", "brands")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()

	// Command should execute successfully (even with empty database)
	// The important thing is that the source flag is accepted
	assert.NoError(t, err) // Should not error with empty database

	stderrOutput := stderr.String()
	// Should not contain flag parsing errors
	assert.NotContains(t, stderrOutput, "unknown flag")
	assert.NotContains(t, stderrOutput, "flag provided but not defined")
}

// Test CLI with verbose flag
func TestCLI_VerboseFlag(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping CLI integration test in short mode")
	}

	binaryPath := buildCLIBinary(t)
	defer os.Remove(binaryPath)

	// Test with verbose flag
	cmd := exec.Command(binaryPath, "--verbose", "print", "brands")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	_ = cmd.Run() // May error due to empty database

	// Command should execute without flag parsing errors
	stderrOutput := stderr.String()
	assert.NotContains(t, stderrOutput, "unknown flag")
	assert.NotContains(t, stderrOutput, "flag provided but not defined")
}