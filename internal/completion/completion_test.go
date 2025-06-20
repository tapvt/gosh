package completion

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"gosh/internal/config"
)

func TestNew(t *testing.T) {
	cfg := config.Default()
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

func TestCompleteCommand(t *testing.T) {
	cfg := config.Default()
	cfg.Aliases["test"] = "echo test"
	mgr, _ := New(cfg)

	tests := []struct {
		name               string
		prefix             string
		expectedBuiltins   []string
		shouldContainAlias bool
	}{
		{
			name:               "prefix 'c'",
			prefix:             "c",
			expectedBuiltins:   []string{"cd"},
			shouldContainAlias: false,
		},
		{
			name:               "prefix 'h'",
			prefix:             "h",
			expectedBuiltins:   []string{"help", "history"},
			shouldContainAlias: false,
		},
		{
			name:               "prefix 'test'",
			prefix:             "test",
			expectedBuiltins:   []string{},
			shouldContainAlias: true,
		},
		{
			name:               "no matches",
			prefix:             "xyz",
			expectedBuiltins:   []string{},
			shouldContainAlias: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			completions, err := mgr.completeCommand(tt.prefix)
			if err != nil {
				t.Errorf("completeCommand() failed: %v", err)
				return
			}

			// Check that all expected builtins are present
			for _, expected := range tt.expectedBuiltins {
				found := false
				for _, completion := range completions {
					if completion == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("completeCommand() missing expected builtin: %s", expected)
				}
			}

			// Check alias presence
			if tt.shouldContainAlias {
				found := false
				for _, completion := range completions {
					if completion == "test" {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("completeCommand() missing expected alias: test")
				}
			}
		})
	}
}

func TestCompleteFile(t *testing.T) {
	// Create a temporary directory structure for testing
	tmpDir := t.TempDir()

	// Create test files and directories
	testFiles := []string{
		"test.txt",
		"test.go",
		"another.txt",
		".hidden",
	}

	testDirs := []string{
		"testdir",
		"anotherdir",
		".hiddendir",
	}

	for _, file := range testFiles {
		f, err := os.Create(filepath.Join(tmpDir, file))
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
		f.Close()
	}

	for _, dir := range testDirs {
		err := os.Mkdir(filepath.Join(tmpDir, dir), 0755)
		if err != nil {
			t.Fatalf("Failed to create test directory %s: %v", dir, err)
		}
	}

	// Change to test directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	cfg := config.Default()
	mgr, _ := New(cfg)

	tests := []struct {
		name     string
		prefix   string
		expected []string
	}{
		{
			name:     "prefix 'test'",
			prefix:   "test",
			expected: []string{"test.go", "test.txt", "testdir/"},
		},
		{
			name:     "prefix 'another'",
			prefix:   "another",
			expected: []string{"another.txt", "anotherdir/"},
		},
		{
			name:     "no matches",
			prefix:   "xyz",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			completions, err := mgr.completeFile(tt.prefix)
			if err != nil {
				t.Errorf("completeFile() failed: %v", err)
				return
			}

			sort.Strings(completions)
			sort.Strings(tt.expected)

			if !reflect.DeepEqual(completions, tt.expected) {
				t.Errorf("completeFile() = %v, want %v", completions, tt.expected)
			}
		})
	}
}

func TestCompleteFileWithHidden(t *testing.T) {
	// Create a temporary directory structure for testing
	tmpDir := t.TempDir()

	// Create test files including hidden ones
	testFiles := []string{
		"visible.txt",
		".hidden.txt",
	}

	for _, file := range testFiles {
		f, err := os.Create(filepath.Join(tmpDir, file))
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
		f.Close()
	}

	// Change to test directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	// Test with hidden files disabled
	cfg := config.Default()
	cfg.CompletionShowHidden = false
	mgr, _ := New(cfg)

	completions, err := mgr.completeFile("")
	if err != nil {
		t.Fatalf("completeFile() failed: %v", err)
	}

	// Should not include hidden files
	for _, completion := range completions {
		if strings.HasPrefix(completion, ".") {
			t.Errorf("completeFile() included hidden file %s when CompletionShowHidden=false", completion)
		}
	}

	// Test with hidden files enabled
	cfg.CompletionShowHidden = true
	mgr, _ = New(cfg)

	completions, err = mgr.completeFile("")
	if err != nil {
		t.Fatalf("completeFile() failed: %v", err)
	}

	// Should include hidden files
	hasHidden := false
	for _, completion := range completions {
		if strings.HasPrefix(completion, ".") {
			hasHidden = true
			break
		}
	}

	if !hasHidden {
		t.Error("completeFile() did not include hidden files when CompletionShowHidden=true")
	}
}

func TestComplete(t *testing.T) {
	cfg := config.Default()
	mgr, _ := New(cfg)

	tests := []struct {
		name      string
		input     string
		cursorPos int
		wantErr   bool
	}{
		{
			name:      "empty input",
			input:     "",
			cursorPos: 0,
			wantErr:   false,
		},
		{
			name:      "command completion",
			input:     "c",
			cursorPos: 1,
			wantErr:   false,
		},
		{
			name:      "file completion",
			input:     "ls test",
			cursorPos: 7,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := mgr.Complete(tt.input, tt.cursorPos)
			if (err != nil) != tt.wantErr {
				t.Errorf("Complete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCompleteGitSubcommands(t *testing.T) {
	cfg := config.Default()
	mgr, _ := New(cfg)

	tests := []struct {
		name     string
		prefix   string
		expected []string
	}{
		{
			name:     "empty prefix",
			prefix:   "",
			expected: []string{"add", "branch", "checkout", "clone", "commit", "diff", "fetch", "init", "log", "merge", "pull", "push", "rebase", "remote", "reset", "show", "status", "switch", "tag"},
		},
		{
			name:     "prefix 'st'",
			prefix:   "st",
			expected: []string{"status"},
		},
		{
			name:     "prefix 'sta'",
			prefix:   "sta",
			expected: []string{"status"},
		},
		{
			name:     "prefix 'sw'",
			prefix:   "sw",
			expected: []string{"switch"},
		},
		{
			name:     "no matches",
			prefix:   "xyz",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			completions, err := mgr.completeGitSubcommands(tt.prefix)
			if err != nil {
				t.Errorf("completeGitSubcommands() failed: %v", err)
				return
			}

			if !reflect.DeepEqual(completions, tt.expected) {
				t.Errorf("completeGitSubcommands() = %v, expected %v", completions, tt.expected)
			}
		})
	}
}

func TestCompleteGit(t *testing.T) {
	cfg := config.Default()
	mgr, _ := New(cfg)

	tests := []struct {
		name      string
		input     string
		cursorPos int
		expected  []string
	}{
		{
			name:      "git st completion",
			input:     "git st",
			cursorPos: 6,
			expected:  []string{"status"},
		},
		{
			name:      "git sta completion",
			input:     "git sta",
			cursorPos: 7,
			expected:  []string{"status"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := strings.Fields(tt.input[:tt.cursorPos])
			completions, err := mgr.completeGit(tokens, tt.cursorPos, tt.input)
			if err != nil {
				t.Errorf("completeGit() failed: %v", err)
				return
			}

			if !reflect.DeepEqual(completions, tt.expected) {
				t.Errorf("completeGit() = %v, expected %v", completions, tt.expected)
			}
		})
	}
}

func TestRemoveDuplicates(t *testing.T) {
	mgr, _ := New(config.Default())

	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "no duplicates",
			input:    []string{"a", "b", "c"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "with duplicates",
			input:    []string{"a", "b", "a", "c", "b"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "empty slice",
			input:    []string{},
			expected: nil,
		},
		{
			name:     "all duplicates",
			input:    []string{"a", "a", "a"},
			expected: []string{"a"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mgr.removeDuplicates(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("removeDuplicates() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetCommonPrefix(t *testing.T) {
	mgr, _ := New(config.Default())

	tests := []struct {
		name        string
		completions []string
		expected    string
	}{
		{
			name:        "empty slice",
			completions: []string{},
			expected:    "",
		},
		{
			name:        "single completion",
			completions: []string{"test"},
			expected:    "test",
		},
		{
			name:        "common prefix",
			completions: []string{"test1", "test2", "test3"},
			expected:    "test",
		},
		{
			name:        "no common prefix",
			completions: []string{"abc", "def", "ghi"},
			expected:    "",
		},
		{
			name:        "partial common prefix",
			completions: []string{"testing", "test", "tester"},
			expected:    "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mgr.GetCommonPrefix(tt.completions)
			if result != tt.expected {
				t.Errorf("GetCommonPrefix() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestCommonPrefix(t *testing.T) {
	mgr, _ := New(config.Default())

	tests := []struct {
		name     string
		a        string
		b        string
		expected string
	}{
		{
			name:     "identical strings",
			a:        "test",
			b:        "test",
			expected: "test",
		},
		{
			name:     "common prefix",
			a:        "testing",
			b:        "tester",
			expected: "test",
		},
		{
			name:     "no common prefix",
			a:        "abc",
			b:        "def",
			expected: "",
		},
		{
			name:     "one empty string",
			a:        "test",
			b:        "",
			expected: "",
		},
		{
			name:     "both empty strings",
			a:        "",
			b:        "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mgr.commonPrefix(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("commonPrefix() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestFormatCompletions(t *testing.T) {
	mgr, _ := New(config.Default())

	tests := []struct {
		name        string
		completions []string
		maxWidth    int
		expected    int // Expected number of formatted lines
	}{
		{
			name:        "empty completions",
			completions: []string{},
			maxWidth:    80,
			expected:    0,
		},
		{
			name:        "single completion",
			completions: []string{"test"},
			maxWidth:    80,
			expected:    1,
		},
		{
			name:        "multiple completions",
			completions: []string{"test1", "test2", "test3", "test4"},
			maxWidth:    20,
			expected:    2, // Should wrap to multiple lines
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mgr.FormatCompletions(tt.completions, tt.maxWidth)
			if len(result) != tt.expected {
				t.Errorf("FormatCompletions() returned %d lines, want %d", len(result), tt.expected)
			}
		})
	}
}
