//go:build integration
// +build integration

// Package main provides integration tests for the gosh shell.
// These tests verify end-to-end functionality and component interactions.
package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestShellBasicCommands tests basic shell functionality
func TestShellBasicCommands(t *testing.T) {
	// Build gosh if not already built
	if err := buildGosh(); err != nil {
		t.Fatalf("Failed to build gosh: %v", err)
	}

	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "pwd command",
			input:    "pwd\nexit\n",
			expected: "/",
			wantErr:  false,
		},
		{
			name:     "help command",
			input:    "help\nexit\n",
			expected: "Gosh - A modern shell written in Go",
			wantErr:  false,
		},
		{
			name:     "alias command",
			input:    "alias\nexit\n",
			expected: "alias ll=",
			wantErr:  false,
		},
		{
			name:     "echo command",
			input:    "echo hello world\nexit\n",
			expected: "hello world",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := runGoshCommand(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("runGoshCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !strings.Contains(output, tt.expected) {
				t.Errorf("runGoshCommand() output = %q, want to contain %q", output, tt.expected)
			}
		})
	}
}

// TestShellConfiguration tests configuration loading and application
func TestShellConfiguration(t *testing.T) {
	if err := buildGosh(); err != nil {
		t.Fatalf("Failed to build gosh: %v", err)
	}

	// Create a temporary config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, ".goshrc")

	configContent := `# Test configuration
alias test_cmd="echo test successful"
export TEST_VAR=integration_test
`

	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Set HOME to temp directory so gosh loads our config
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tmpDir)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "alias from config",
			input:    "test_cmd\nexit\n",
			expected: "test successful",
		},
		{
			name:     "environment variable from config",
			input:    "echo $TEST_VAR\nexit\n",
			expected: "integration_test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := runGoshCommand(tt.input)
			if err != nil {
				t.Errorf("runGoshCommand() error = %v", err)
				return
			}

			if !strings.Contains(output, tt.expected) {
				t.Errorf("runGoshCommand() output = %q, want to contain %q", output, tt.expected)
			}
		})
	}
}

// TestShellHistory tests command history functionality
func TestShellHistory(t *testing.T) {
	if err := buildGosh(); err != nil {
		t.Fatalf("Failed to build gosh: %v", err)
	}

	// Create a temporary directory for history
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, ".gosh_history")

	// Set environment variables for history
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tmpDir)

	// First session: add some commands to history
	input1 := "echo first command\necho second command\nexit\n"
	_, err := runGoshCommand(input1)
	if err != nil {
		t.Fatalf("First session failed: %v", err)
	}

	// Check if history file was created
	if _, err := os.Stat(historyFile); os.IsNotExist(err) {
		t.Skip("History file not created, skipping history test")
	}

	// Second session: check history
	input2 := "history\nexit\n"
	output, err := runGoshCommand(input2)
	if err != nil {
		t.Fatalf("Second session failed: %v", err)
	}

	// Should contain commands from first session
	if !strings.Contains(output, "first command") {
		t.Errorf("History should contain 'first command', got: %s", output)
	}
}

// TestShellDirectoryNavigation tests cd and pwd commands
func TestShellDirectoryNavigation(t *testing.T) {
	if err := buildGosh(); err != nil {
		t.Fatalf("Failed to build gosh: %v", err)
	}

	// Create a temporary directory structure
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "testdir")
	err := os.Mkdir(subDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Test cd and pwd
	input := fmt.Sprintf("cd %s\npwd\nexit\n", subDir)
	output, err := runGoshCommand(input)
	if err != nil {
		t.Fatalf("runGoshCommand() error = %v", err)
	}

	if !strings.Contains(output, "testdir") {
		t.Errorf("Expected output to contain 'testdir', got: %s", output)
	}
}

// TestShellErrorHandling tests error handling and recovery
func TestShellErrorHandling(t *testing.T) {
	if err := buildGosh(); err != nil {
		t.Fatalf("Failed to build gosh: %v", err)
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "non-existent command",
			input:    "nonexistentcommand123\nexit\n",
			expected: "command not found",
		},
		{
			name:     "invalid cd",
			input:    "cd /nonexistent/directory\nexit\n",
			expected: "no such file or directory",
		},
		{
			name:     "continue after error",
			input:    "nonexistentcommand123\necho still working\nexit\n",
			expected: "still working",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := runGoshCommand(tt.input)
			// We expect the command to succeed even if individual commands fail
			if err != nil {
				t.Errorf("runGoshCommand() should not fail on command errors: %v", err)
				return
			}

			if !strings.Contains(strings.ToLower(output), strings.ToLower(tt.expected)) {
				t.Errorf("runGoshCommand() output = %q, want to contain %q", output, tt.expected)
			}
		})
	}
}

// TestShellVersionAndHelp tests version and help flags
func TestShellVersionAndHelp(t *testing.T) {
	if err := buildGosh(); err != nil {
		t.Fatalf("Failed to build gosh: %v", err)
	}

	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "version flag",
			args:     []string{"--version"},
			expected: "gosh version",
		},
		{
			name:     "help flag",
			args:     []string{"--help"},
			expected: "Usage:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("./build/gosh", tt.args...)
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Errorf("Command failed: %v, output: %s", err, output)
				return
			}

			if !strings.Contains(string(output), tt.expected) {
				t.Errorf("Output = %q, want to contain %q", string(output), tt.expected)
			}
		})
	}
}

// TestShellInteractiveFeatures tests interactive features
func TestShellInteractiveFeatures(t *testing.T) {
	if err := buildGosh(); err != nil {
		t.Fatalf("Failed to build gosh: %v", err)
	}

	// Test that shell starts and can handle basic interaction
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "./build/gosh")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("Failed to create stdin pipe: %v", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("Failed to create stdout pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start gosh: %v", err)
	}

	// Send a simple command
	go func() {
		defer stdin.Close()
		fmt.Fprintln(stdin, "echo interactive test")
		fmt.Fprintln(stdin, "exit")
	}()

	// Read output
	scanner := bufio.NewScanner(stdout)
	var output strings.Builder
	for scanner.Scan() {
		output.WriteString(scanner.Text() + "\n")
	}

	if err := cmd.Wait(); err != nil {
		t.Errorf("Command failed: %v", err)
	}

	outputStr := output.String()
	if !strings.Contains(outputStr, "interactive test") {
		t.Errorf("Expected output to contain 'interactive test', got: %s", outputStr)
	}
}

// Helper functions

// buildGosh builds the gosh binary if it doesn't exist
func buildGosh() error {
	if _, err := os.Stat("./build/gosh"); err == nil {
		return nil // Already built
	}

	cmd := exec.Command("make", "build")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("build failed: %v, output: %s", err, output)
	}
	return nil
}

// runGoshCommand runs gosh with the given input and returns the output
func runGoshCommand(input string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "./build/gosh")
	cmd.Stdin = strings.NewReader(input)

	output, err := cmd.CombinedOutput()
	return string(output), err
}
