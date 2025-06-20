package shell

import (
	"testing"

	"gosh/internal/completion"
	"gosh/internal/config"
)

func TestShellCompleter_Do(t *testing.T) {
	cfg := config.Default()
	completionMgr, err := completion.New(cfg)
	if err != nil {
		t.Fatalf("Failed to create completion manager: %v", err)
	}

	completer := &shellCompleter{completion: completionMgr}

	tests := []struct {
		name           string
		line           string
		pos            int
		expectedLength int
		description    string
	}{
		{
			name:           "git st completion",
			line:           "git st",
			pos:            6,
			expectedLength: 2, // Should replace "st" (2 characters)
			description:    "Should replace 'st' with 'status', not create 'ststatus'",
		},
		{
			name:           "git sta completion",
			line:           "git sta",
			pos:            7,
			expectedLength: 3, // Should replace "sta" (3 characters)
			description:    "Should replace 'sta' with 'status'",
		},
		{
			name:           "git s completion",
			line:           "git s",
			pos:            5,
			expectedLength: 1, // Should replace "s" (1 character)
			description:    "Should replace 's' with common prefix of 'show', 'status', 'switch'",
		},
		{
			name:           "command at start",
			line:           "h",
			pos:            1,
			expectedLength: 1, // Should replace "h" (1 character)
			description:    "Should complete command at start of line",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lineRunes := []rune(tt.line)
			completions, length := completer.Do(lineRunes, tt.pos)

			// Check that we got some completions
			if len(completions) == 0 {
				t.Errorf("Expected completions but got none")
				return
			}

			// Check that the length is correct
			if length != tt.expectedLength {
				t.Errorf("Expected length %d, got %d. %s", tt.expectedLength, length, tt.description)
			}

			t.Logf("Test %s: line=%q, pos=%d, completions=%v, length=%d",
				tt.name, tt.line, tt.pos, completionsToStrings(completions), length)
		})
	}
}

func TestShellCompleter_GitStCompletion(t *testing.T) {
	// This test specifically verifies the fix for the "git st" -> "git ststatus" bug
	cfg := config.Default()
	completionMgr, err := completion.New(cfg)
	if err != nil {
		t.Fatalf("Failed to create completion manager: %v", err)
	}

	completer := &shellCompleter{completion: completionMgr}

	// Test the exact scenario that was failing
	line := "git st"
	pos := 6 // cursor is at the end of "git st"
	lineRunes := []rune(line)

	completions, length := completer.Do(lineRunes, pos)

	// Should get "status" as a completion
	if len(completions) == 0 {
		t.Fatal("Expected completions but got none")
	}

	// Should replace exactly 2 characters ("st")
	if length != 2 {
		t.Errorf("Expected to replace 2 characters ('st'), but got length %d", length)
		t.Errorf("This would cause 'git st' + 'status' to become 'git ststatus' instead of 'git status'")
	}

	// Verify we got "atus" (the suffix to complete "st" -> "status") as one of the completions
	found := false
	for _, completion := range completions {
		if string(completion) == "atus" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected 'atus' (suffix for 'st' -> 'status') in completions, got: %v", completionsToStrings(completions))
	}
}

// Helper function to convert [][]rune to []string for easier testing
func completionsToStrings(completions [][]rune) []string {
	var result []string
	for _, completion := range completions {
		result = append(result, string(completion))
	}
	return result
}
