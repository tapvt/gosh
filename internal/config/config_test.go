package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()

	// Test default values
	if cfg.Debug != false {
		t.Errorf("Expected Debug to be false, got %v", cfg.Debug)
	}

	if cfg.ShowWelcome != true {
		t.Errorf("Expected ShowWelcome to be true, got %v", cfg.ShowWelcome)
	}

	if cfg.HistorySize != 10000 {
		t.Errorf("Expected HistorySize to be 10000, got %v", cfg.HistorySize)
	}

	if cfg.CompletionEnabled != true {
		t.Errorf("Expected CompletionEnabled to be true, got %v", cfg.CompletionEnabled)
	}

	if cfg.GitEnabled != true {
		t.Errorf("Expected GitEnabled to be true, got %v", cfg.GitEnabled)
	}

	// Test default aliases
	if cfg.Aliases["ll"] != "ls -la" {
		t.Errorf("Expected alias 'll' to be 'ls -la', got %v", cfg.Aliases["ll"])
	}

	// Test environment map is initialized
	if cfg.Environment == nil {
		t.Error("Expected Environment map to be initialized")
	}
}

func TestParseBool(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"True", true},
		{"TRUE", true},
		{"yes", true},
		{"Yes", true},
		{"YES", true},
		{"1", true},
		{"on", true},
		{"On", true},
		{"ON", true},
		{"false", false},
		{"False", false},
		{"FALSE", false},
		{"no", false},
		{"0", false},
		{"off", false},
		{"", false},
		{"invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseBool(tt.input)
			if result != tt.expected {
				t.Errorf("parseBool(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseExport(t *testing.T) {
	cfg := Default()

	tests := []struct {
		name     string
		input    string
		wantErr  bool
		checkKey string
		checkVal string
	}{
		{
			name:     "simple export",
			input:    "PATH=/usr/bin",
			wantErr:  false,
			checkKey: "PATH",
			checkVal: "/usr/bin",
		},
		{
			name:     "quoted export",
			input:    `EDITOR="vim"`,
			wantErr:  false,
			checkKey: "EDITOR",
			checkVal: "vim",
		},
		{
			name:     "single quoted export",
			input:    "SHELL='/bin/bash'",
			wantErr:  false,
			checkKey: "SHELL",
			checkVal: "/bin/bash",
		},
		{
			name:    "invalid export",
			input:   "INVALID",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cfg.parseExport(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseExport() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if val, ok := cfg.Environment[tt.checkKey]; !ok || val != tt.checkVal {
					t.Errorf("Expected Environment[%s] = %s, got %s", tt.checkKey, tt.checkVal, val)
				}
			}
		})
	}
}

func TestParseAlias(t *testing.T) {
	cfg := Default()

	tests := []struct {
		name     string
		input    string
		wantErr  bool
		checkKey string
		checkVal string
	}{
		{
			name:     "simple alias",
			input:    "ls=ls --color=auto",
			wantErr:  false,
			checkKey: "ls",
			checkVal: "ls --color=auto",
		},
		{
			name:     "quoted alias",
			input:    `ll="ls -la"`,
			wantErr:  false,
			checkKey: "ll",
			checkVal: "ls -la",
		},
		{
			name:    "invalid alias",
			input:   "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cfg.parseAlias(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseAlias() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if val, ok := cfg.Aliases[tt.checkKey]; !ok || val != tt.checkVal {
					t.Errorf("Expected Aliases[%s] = %s, got %s", tt.checkKey, tt.checkVal, val)
				}
			}
		})
	}
}

func TestSetConfigValue(t *testing.T) {
	cfg := Default()

	tests := []struct {
		name    string
		key     string
		value   string
		wantErr bool
		check   func(*Config) bool
	}{
		{
			name:    "set debug",
			key:     "DEBUG",
			value:   "true",
			wantErr: false,
			check:   func(c *Config) bool { return c.Debug == true },
		},
		{
			name:    "set history size",
			key:     "HISTORY_SIZE",
			value:   "5000",
			wantErr: false,
			check:   func(c *Config) bool { return c.HistorySize == 5000 },
		},
		{
			name:    "set prompt format",
			key:     "PROMPT_FORMAT",
			value:   "%u$ ",
			wantErr: false,
			check:   func(c *Config) bool { return c.PromptFormat == "%u$ " },
		},
		{
			name:    "unknown key",
			key:     "UNKNOWN_KEY",
			value:   "value",
			wantErr: true,
			check:   func(c *Config) bool { return true },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cfg.setConfigValue(tt.key, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("setConfigValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !tt.check(cfg) {
				t.Errorf("setConfigValue() did not set value correctly")
			}
		})
	}
}

func TestLoadFromFile(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "test_config")

	configContent := `# Test configuration
export TEST_VAR=test_value
alias test_alias="echo test"
set DEBUG=true
GOSH_HISTORY_SIZE=5000
`

	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	cfg := Default()
	err = cfg.loadFromFile(configFile)
	if err != nil {
		t.Fatalf("loadFromFile() failed: %v", err)
	}

	// Check that values were loaded correctly
	if cfg.Environment["TEST_VAR"] != "test_value" {
		t.Errorf("Expected TEST_VAR=test_value, got %s", cfg.Environment["TEST_VAR"])
	}

	if cfg.Aliases["test_alias"] != "echo test" {
		t.Errorf("Expected test_alias='echo test', got %s", cfg.Aliases["test_alias"])
	}

	if cfg.Debug != true {
		t.Errorf("Expected Debug=true, got %v", cfg.Debug)
	}

	if cfg.HistorySize != 5000 {
		t.Errorf("Expected HistorySize=5000, got %v", cfg.HistorySize)
	}
}

func TestParseLine(t *testing.T) {
	cfg := Default()

	tests := []struct {
		name    string
		line    string
		wantErr bool
	}{
		{
			name:    "comment line",
			line:    "# This is a comment",
			wantErr: false,
		},
		{
			name:    "empty line",
			line:    "",
			wantErr: false,
		},
		{
			name:    "export statement",
			line:    "export PATH=/usr/bin",
			wantErr: false,
		},
		{
			name:    "alias statement",
			line:    "alias ll='ls -la'",
			wantErr: false,
		},
		{
			name:    "set statement",
			line:    "set DEBUG=true",
			wantErr: false,
		},
		{
			name:    "assignment",
			line:    "GOSH_PROMPT_FORMAT='%u$ '",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cfg.parseLine(tt.line)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseLine() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoad_NonExistentFile(t *testing.T) {
	tmpDir := t.TempDir()
	nonExistentPath := filepath.Join(tmpDir, "nonexistent")

	_, err := Load(nonExistentPath)
	if !os.IsNotExist(err) {
		t.Errorf("Expected os.IsNotExist error, got %v", err)
	}
}

func TestParseAssignment(t *testing.T) {
	cfg := Default()

	tests := []struct {
		name    string
		line    string
		wantErr bool
	}{
		{
			name:    "GOSH_ prefixed variable",
			line:    "GOSH_DEBUG=true",
			wantErr: false,
		},
		{
			name:    "regular environment variable",
			line:    "PATH=/usr/bin",
			wantErr: false,
		},
		{
			name:    "invalid assignment",
			line:    "INVALID",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cfg.parseAssignment(tt.line)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseAssignment() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
