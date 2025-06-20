package history

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"gosh/internal/config"
)

func TestNew(t *testing.T) {
	cfg := config.Default()
	cfg.SaveHistory = false // Don't try to load from file

	mgr, err := New(cfg)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	if mgr == nil {
		t.Fatal("New() returned nil manager")
	}
	if mgr.config != cfg {
		t.Error("New() did not set config correctly")
	}
}

func TestAdd(t *testing.T) {
	cfg := config.Default()
	cfg.SaveHistory = false
	cfg.HistorySize = 5
	mgr, _ := New(cfg)

	tests := []struct {
		name     string
		commands []string
		expected []string
	}{
		{
			name:     "add single command",
			commands: []string{"ls"},
			expected: []string{"ls"},
		},
		{
			name:     "add multiple commands",
			commands: []string{"ls", "pwd", "cd"},
			expected: []string{"ls", "pwd", "cd"},
		},
		{
			name:     "exceed history size",
			commands: []string{"cmd1", "cmd2", "cmd3", "cmd4", "cmd5", "cmd6"},
			expected: []string{"cmd2", "cmd3", "cmd4", "cmd5", "cmd6"},
		},
		{
			name:     "empty command",
			commands: []string{"ls", "", "pwd"},
			expected: []string{"ls", "pwd"},
		},
		{
			name:     "whitespace only command",
			commands: []string{"ls", "   ", "pwd"},
			expected: []string{"ls", "pwd"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset manager
			mgr, _ = New(cfg)

			for _, cmd := range tt.commands {
				mgr.Add(cmd)
			}

			entries := mgr.GetAll()
			var commands []string
			for _, entry := range entries {
				commands = append(commands, entry.Command)
			}

			if !reflect.DeepEqual(commands, tt.expected) {
				t.Errorf("Add() resulted in %v, want %v", commands, tt.expected)
			}
		})
	}
}

func TestAddWithDuplicates(t *testing.T) {
	// Test with duplicates disabled
	cfg := config.Default()
	cfg.SaveHistory = false
	cfg.HistoryDuplicates = false
	mgr, _ := New(cfg)

	mgr.Add("ls")
	mgr.Add("pwd")
	mgr.Add("pwd") // Consecutive duplicate - should be skipped
	mgr.Add("cd")

	entries := mgr.GetAll()
	var commands []string
	for _, entry := range entries {
		commands = append(commands, entry.Command)
	}

	// With duplicates disabled, only consecutive duplicates are prevented
	expected := []string{"ls", "pwd", "cd"}
	if !reflect.DeepEqual(commands, expected) {
		t.Errorf("Add() with duplicates disabled resulted in %v, want %v", commands, expected)
	}

	// Test with duplicates enabled
	cfg2 := config.Default()
	cfg2.SaveHistory = false
	cfg2.HistoryDuplicates = true
	mgr2, _ := New(cfg2)

	mgr2.Add("ls")
	mgr2.Add("pwd")
	mgr2.Add("pwd") // Consecutive duplicate should be kept
	mgr2.Add("cd")

	entries2 := mgr2.GetAll()
	commands2 := []string{}
	for _, entry := range entries2 {
		commands2 = append(commands2, entry.Command)
	}

	expected2 := []string{"ls", "pwd", "pwd", "cd"}
	if !reflect.DeepEqual(commands2, expected2) {
		t.Errorf("Add() with duplicates enabled resulted in %v, want %v", commands2, expected2)
	}
}

func TestGetRecent(t *testing.T) {
	cfg := config.Default()
	cfg.SaveHistory = false
	mgr, _ := New(cfg)

	commands := []string{"cmd1", "cmd2", "cmd3", "cmd4", "cmd5"}
	for _, cmd := range commands {
		mgr.Add(cmd)
	}

	tests := []struct {
		name     string
		n        int
		expected []string
	}{
		{
			name:     "get last 3",
			n:        3,
			expected: []string{"cmd3", "cmd4", "cmd5"},
		},
		{
			name:     "get more than available",
			n:        10,
			expected: []string{"cmd1", "cmd2", "cmd3", "cmd4", "cmd5"},
		},
		{
			name:     "get zero",
			n:        0,
			expected: nil,
		},
		{
			name:     "get negative",
			n:        -1,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entries := mgr.GetRecent(tt.n)
			var commands []string
			for _, entry := range entries {
				commands = append(commands, entry.Command)
			}

			if !reflect.DeepEqual(commands, tt.expected) {
				t.Errorf("GetRecent(%d) = %v, want %v", tt.n, commands, tt.expected)
			}
		})
	}
}

func TestSearch(t *testing.T) {
	cfg := config.Default()
	cfg.SaveHistory = false
	mgr, _ := New(cfg)

	commands := []string{"ls -la", "git status", "git commit", "pwd", "ls"}
	for _, cmd := range commands {
		mgr.Add(cmd)
	}

	tests := []struct {
		name     string
		term     string
		expected []string
	}{
		{
			name:     "search for 'git'",
			term:     "git",
			expected: []string{"git status", "git commit"},
		},
		{
			name:     "search for 'ls'",
			term:     "ls",
			expected: []string{"ls -la", "ls"},
		},
		{
			name:     "search for 'status'",
			term:     "status",
			expected: []string{"git status"},
		},
		{
			name:     "search for non-existent",
			term:     "nonexistent",
			expected: nil,
		},
		{
			name:     "empty search term",
			term:     "",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entries := mgr.Search(tt.term)
			var commands []string
			for _, entry := range entries {
				commands = append(commands, entry.Command)
			}

			if !reflect.DeepEqual(commands, tt.expected) {
				t.Errorf("Search(%q) = %v, want %v", tt.term, commands, tt.expected)
			}
		})
	}
}

func TestSearchPrefix(t *testing.T) {
	cfg := config.Default()
	cfg.SaveHistory = false
	mgr, _ := New(cfg)

	commands := []string{"ls -la", "git status", "git commit", "pwd", "ls"}
	for _, cmd := range commands {
		mgr.Add(cmd)
	}

	tests := []struct {
		name     string
		prefix   string
		expected []string
	}{
		{
			name:     "search prefix 'git'",
			prefix:   "git",
			expected: []string{"git status", "git commit"},
		},
		{
			name:     "search prefix 'ls'",
			prefix:   "ls",
			expected: []string{"ls -la", "ls"},
		},
		{
			name:     "search prefix 'p'",
			prefix:   "p",
			expected: []string{"pwd"},
		},
		{
			name:     "search prefix 'xyz'",
			prefix:   "xyz",
			expected: nil,
		},
		{
			name:     "empty prefix",
			prefix:   "",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entries := mgr.SearchPrefix(tt.prefix)
			var commands []string
			for _, entry := range entries {
				commands = append(commands, entry.Command)
			}

			if !reflect.DeepEqual(commands, tt.expected) {
				t.Errorf("SearchPrefix(%q) = %v, want %v", tt.prefix, commands, tt.expected)
			}
		})
	}
}

func TestNavigation(t *testing.T) {
	cfg := config.Default()
	cfg.SaveHistory = false
	mgr, _ := New(cfg)

	commands := []string{"cmd1", "cmd2", "cmd3"}
	for _, cmd := range commands {
		mgr.Add(cmd)
	}

	// Test Previous navigation
	tests := []struct {
		name     string
		action   string
		expected string
	}{
		{"first previous", "previous", "cmd3"},
		{"second previous", "previous", "cmd2"},
		{"third previous", "previous", "cmd1"},
		{"beyond start", "previous", "cmd1"},
		{"next from start", "next", "cmd2"},
		{"next again", "next", "cmd3"},
		{"beyond end", "next", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result string
			if tt.action == "previous" {
				result = mgr.Previous()
			} else {
				result = mgr.Next()
			}

			if result != tt.expected {
				t.Errorf("%s() = %q, want %q", tt.action, result, tt.expected)
			}
		})
	}
}

func TestClear(t *testing.T) {
	cfg := config.Default()
	cfg.SaveHistory = false
	mgr, _ := New(cfg)

	mgr.Add("cmd1")
	mgr.Add("cmd2")

	if len(mgr.GetAll()) != 2 {
		t.Error("Expected 2 entries before clear")
	}

	err := mgr.Clear()
	if err != nil {
		t.Errorf("Clear() failed: %v", err)
	}

	if len(mgr.GetAll()) != 0 {
		t.Error("Expected 0 entries after clear")
	}
}

func TestSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "test_history")

	// Create manager with save enabled
	cfg := config.Default()
	cfg.SaveHistory = true
	cfg.HistoryFile = historyFile
	mgr, err := New(cfg)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	// Add some commands
	commands := []string{"ls", "pwd", "git status"}
	for _, cmd := range commands {
		mgr.Add(cmd)
	}

	// Create a new manager to load the history
	mgr2, err := New(cfg)
	if err != nil {
		t.Fatalf("New() for loading failed: %v", err)
	}

	entries := mgr2.GetAll()
	var loadedCommands []string
	for _, entry := range entries {
		loadedCommands = append(loadedCommands, entry.Command)
	}

	if !reflect.DeepEqual(loadedCommands, commands) {
		t.Errorf("Loaded commands %v, want %v", loadedCommands, commands)
	}
}

func TestGetStats(t *testing.T) {
	cfg := config.Default()
	cfg.SaveHistory = false
	mgr, _ := New(cfg)

	mgr.Add("ls")
	mgr.Add("pwd")
	mgr.Add("ls") // Duplicate

	stats := mgr.GetStats()

	if stats["total_entries"] != 3 {
		t.Errorf("Expected total_entries=3, got %v", stats["total_entries"])
	}

	if stats["unique_commands"] != 2 {
		t.Errorf("Expected unique_commands=2, got %v", stats["unique_commands"])
	}

	if stats["max_size"] != cfg.HistorySize {
		t.Errorf("Expected max_size=%d, got %v", cfg.HistorySize, stats["max_size"])
	}
}

func TestExport(t *testing.T) {
	cfg := config.Default()
	cfg.SaveHistory = false
	mgr, _ := New(cfg)

	mgr.Add("ls")
	mgr.Add("pwd")

	tmpDir := t.TempDir()

	// Test bash export
	bashFile := filepath.Join(tmpDir, "history.bash")
	err := mgr.Export(bashFile, "bash")
	if err != nil {
		t.Errorf("Export to bash format failed: %v", err)
	}

	content, err := os.ReadFile(bashFile)
	if err != nil {
		t.Errorf("Failed to read exported bash file: %v", err)
	}

	expectedLines := []string{"ls", "pwd"}
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if !reflect.DeepEqual(lines, expectedLines) {
		t.Errorf("Bash export content = %v, want %v", lines, expectedLines)
	}

	// Test JSON export
	jsonFile := filepath.Join(tmpDir, "history.json")
	err = mgr.Export(jsonFile, "json")
	if err != nil {
		t.Errorf("Export to JSON format failed: %v", err)
	}

	content, err = os.ReadFile(jsonFile)
	if err != nil {
		t.Errorf("Failed to read exported JSON file: %v", err)
	}

	if !strings.Contains(string(content), `"command": "ls"`) {
		t.Error("JSON export does not contain expected command")
	}

	// Test unsupported format
	err = mgr.Export(filepath.Join(tmpDir, "history.xml"), "xml")
	if err == nil {
		t.Error("Expected error for unsupported export format")
	}
}

func TestReset(t *testing.T) {
	cfg := config.Default()
	cfg.SaveHistory = false
	mgr, _ := New(cfg)

	mgr.Add("cmd1")
	mgr.Add("cmd2")

	// Navigate in history
	mgr.Previous()
	mgr.Previous()

	// Reset should set current to end
	mgr.Reset()

	// Next should return empty (at end)
	result := mgr.Next()
	if result != "" {
		t.Errorf("After reset, Next() = %q, want empty string", result)
	}
}
