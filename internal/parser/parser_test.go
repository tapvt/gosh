package parser

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gosh/internal/config"
)

func TestTokenize(t *testing.T) {
	parser := New(config.Default())

	tests := []struct {
		name     string
		input    string
		expected []string
		wantErr  bool
	}{
		{
			name:     "simple command",
			input:    "ls -la",
			expected: []string{"ls", "-la"},
			wantErr:  false,
		},
		{
			name:     "quoted arguments",
			input:    `echo "hello world"`,
			expected: []string{"echo", "hello world"},
			wantErr:  false,
		},
		{
			name:     "single quoted arguments",
			input:    "echo 'hello world'",
			expected: []string{"echo", "hello world"},
			wantErr:  false,
		},
		{
			name:     "mixed quotes",
			input:    `echo "hello" 'world'`,
			expected: []string{"echo", "hello", "world"},
			wantErr:  false,
		},
		{
			name:     "escaped characters",
			input:    `echo hello\ world`,
			expected: []string{"echo", "hello world"},
			wantErr:  false,
		},
		{
			name:     "empty input",
			input:    "",
			expected: []string{},
			wantErr:  false,
		},
		{
			name:     "whitespace only",
			input:    "   \t  ",
			expected: []string{},
			wantErr:  false,
		},
		{
			name:    "unclosed quote",
			input:   `echo "unclosed`,
			wantErr: true,
		},
		{
			name:     "escaped quote",
			input:    `echo "hello \"world\""`,
			expected: []string{"echo", `hello "world"`},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens, err := parser.tokenize(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("tokenize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(tokens) != len(tt.expected) {
					t.Errorf("tokenize() got %d tokens, want %d", len(tokens), len(tt.expected))
					return
				}

				for i, token := range tokens {
					if token != tt.expected[i] {
						t.Errorf("tokenize() token[%d] = %q, want %q", i, token, tt.expected[i])
					}
				}
			}
		})
	}
}

func TestExpandAlias(t *testing.T) {
	cfg := config.Default()
	cfg.Aliases["ll"] = "ls -la"
	cfg.Aliases["gs"] = "git status"

	parser := New(cfg)

	tests := []struct {
		name     string
		input    string
		expected string
		expanded bool
	}{
		{
			name:     "expand ll alias",
			input:    "ll /home",
			expected: "ls -la /home",
			expanded: true,
		},
		{
			name:     "expand gs alias",
			input:    "gs",
			expected: "git status",
			expanded: true,
		},
		{
			name:     "no alias",
			input:    "ls -la",
			expected: "ls -la",
			expanded: false,
		},
		{
			name:     "empty input",
			input:    "",
			expected: "",
			expanded: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, expanded := parser.expandAlias(tt.input)
			if result != tt.expected {
				t.Errorf("expandAlias() result = %q, want %q", result, tt.expected)
			}
			if expanded != tt.expanded {
				t.Errorf("expandAlias() expanded = %v, want %v", expanded, tt.expanded)
			}
		})
	}
}

func TestParseBuiltin(t *testing.T) {
	parser := New(config.Default())

	tests := []struct {
		name      string
		tokens    []string
		isBuiltin bool
	}{
		{
			name:      "cd command",
			tokens:    []string{"cd", "/home"},
			isBuiltin: true,
		},
		{
			name:      "pwd command",
			tokens:    []string{"pwd"},
			isBuiltin: true,
		},
		{
			name:      "exit command",
			tokens:    []string{"exit"},
			isBuiltin: true,
		},
		{
			name:      "help command",
			tokens:    []string{"help"},
			isBuiltin: true,
		},
		{
			name:      "history command",
			tokens:    []string{"history"},
			isBuiltin: true,
		},
		{
			name:      "alias command",
			tokens:    []string{"alias"},
			isBuiltin: true,
		},
		{
			name:      "export command",
			tokens:    []string{"export"},
			isBuiltin: true,
		},
		{
			name:      "non-builtin command",
			tokens:    []string{"ls"},
			isBuiltin: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := parser.parseBuiltin(tt.tokens)

			if tt.isBuiltin && cmd == nil {
				t.Errorf("parseBuiltin() returned nil for builtin command %s", tt.tokens[0])
			}

			if !tt.isBuiltin && cmd != nil {
				t.Errorf("parseBuiltin() returned non-nil for non-builtin command %s", tt.tokens[0])
			}
		})
	}
}

func TestCdCommand(t *testing.T) {
	// Save current directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "cd to temp directory",
			args:    []string{tmpDir},
			wantErr: false,
		},
		{
			name:    "cd to home directory",
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "cd to non-existent directory",
			args:    []string{"/non/existent/path"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &CdCommand{Args: tt.args}
			err := cmd.Execute(context.Background(), config.Default())

			if (err != nil) != tt.wantErr {
				t.Errorf("CdCommand.Execute() error = %v, wantErr %v", err, tt.wantErr)
			}

			// If successful and we specified a directory, check we're there
			if !tt.wantErr && len(tt.args) > 0 {
				currentDir, err := os.Getwd()
				if err != nil {
					t.Errorf("Failed to get current directory after cd: %v", err)
				}

				expectedDir := tt.args[0]
				if expectedDir != tmpDir {
					// For home directory test, just check that we changed directories
					return
				}

				// Resolve both paths to handle symlinks
				currentResolved, _ := filepath.EvalSymlinks(currentDir)
				expectedResolved, _ := filepath.EvalSymlinks(expectedDir)

				if currentResolved != expectedResolved {
					t.Errorf("Expected to be in %s, but in %s", expectedResolved, currentResolved)
				}
			}
		})
	}
}

func TestCdCommandTildeExpansion(t *testing.T) {
	// Save current directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Skip("Cannot get home directory")
	}

	// Create a test subdirectory in home
	testDir := filepath.Join(homeDir, "gosh_test_dir")
	os.Mkdir(testDir, 0755)
	defer os.RemoveAll(testDir)

	cmd := &CdCommand{Args: []string{"~/gosh_test_dir"}}
	err = cmd.Execute(context.Background(), config.Default())
	if err != nil {
		t.Errorf("CdCommand.Execute() with tilde expansion failed: %v", err)
	}

	currentDir, err := os.Getwd()
	if err != nil {
		t.Errorf("Failed to get current directory: %v", err)
	}

	if currentDir != testDir {
		t.Errorf("Expected to be in %s, but in %s", testDir, currentDir)
	}
}

func TestPwdCommand(t *testing.T) {
	cmd := &PwdCommand{}
	err := cmd.Execute(context.Background(), config.Default())
	if err != nil {
		t.Errorf("PwdCommand.Execute() failed: %v", err)
	}
}

func TestHelpCommand(t *testing.T) {
	cmd := &HelpCommand{Args: []string{}}
	err := cmd.Execute(context.Background(), config.Default())
	if err != nil {
		t.Errorf("HelpCommand.Execute() failed: %v", err)
	}
}

func TestAliasCommand(t *testing.T) {
	cfg := config.Default()

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "show all aliases",
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "create new alias",
			args:    []string{"test=echo hello"},
			wantErr: false,
		},
		{
			name:    "invalid alias format",
			args:    []string{"invalid"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &AliasCommand{Args: tt.args, Config: cfg}
			err := cmd.Execute(context.Background(), cfg)

			if (err != nil) != tt.wantErr {
				t.Errorf("AliasCommand.Execute() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Check if alias was created
			if !tt.wantErr && len(tt.args) > 0 && strings.Contains(tt.args[0], "=") {
				parts := strings.SplitN(tt.args[0], "=", 2)
				aliasName := parts[0]
				expectedValue := parts[1]

				if cfg.Aliases[aliasName] != expectedValue {
					t.Errorf("Expected alias %s=%s, got %s", aliasName, expectedValue, cfg.Aliases[aliasName])
				}
			}
		})
	}
}

func TestExportCommand(t *testing.T) {
	cfg := config.Default()

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "show all exports",
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "create new export",
			args:    []string{"TEST_VAR=test_value"},
			wantErr: false,
		},
		{
			name:    "invalid export format",
			args:    []string{"invalid"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &ExportCommand{Args: tt.args, Config: cfg}
			err := cmd.Execute(context.Background(), cfg)

			if (err != nil) != tt.wantErr {
				t.Errorf("ExportCommand.Execute() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Check if environment variable was set
			if !tt.wantErr && len(tt.args) > 0 && strings.Contains(tt.args[0], "=") {
				parts := strings.SplitN(tt.args[0], "=", 2)
				varName := parts[0]
				expectedValue := parts[1]

				if cfg.Environment[varName] != expectedValue {
					t.Errorf("Expected environment %s=%s, got %s", varName, expectedValue, cfg.Environment[varName])
				}

				// Check if it was set in the actual environment
				if os.Getenv(varName) != expectedValue {
					t.Errorf("Expected os environment %s=%s, got %s", varName, expectedValue, os.Getenv(varName))
				}
			}
		})
	}
}

func TestNoOpCommand(t *testing.T) {
	cmd := &NoOpCommand{}
	err := cmd.Execute(context.Background(), config.Default())
	if err != nil {
		t.Errorf("NoOpCommand.Execute() failed: %v", err)
	}
}

func TestParse(t *testing.T) {
	parser := New(config.Default())

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "empty input",
			input:   "",
			wantErr: false,
		},
		{
			name:    "builtin command",
			input:   "pwd",
			wantErr: false,
		},
		{
			name:    "external command",
			input:   "ls -la",
			wantErr: false,
		},
		{
			name:    "command with alias",
			input:   "ll", // Should expand to "ls -la"
			wantErr: false,
		},
		{
			name:    "invalid quote",
			input:   `echo "unclosed`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := parser.Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && cmd == nil {
				t.Error("Parse() returned nil command without error")
			}
		})
	}
}
