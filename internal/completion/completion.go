// Package completion provides tab completion functionality for gosh.
// It implements intelligent completion for commands, files, directories,
// and context-aware suggestions.
package completion

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gosh/internal/config"
)

// Manager handles tab completion functionality
type Manager struct {
	config *config.Config
}

// New creates a new completion manager
func New(cfg *config.Config) (*Manager, error) {
	return &Manager{
		config: cfg,
	}, nil
}

// Complete provides completions for the given input
func (m *Manager) Complete(input string, cursorPos int) ([]string, error) {
	if !m.config.CompletionEnabled {
		return nil, nil
	}

	// Parse the input to understand context
	tokens := strings.Fields(input[:cursorPos])
	if len(tokens) == 0 {
		return m.completeCommand("")
	}

	// If we're at the beginning or completing the first token, complete commands
	if len(tokens) == 1 && !strings.HasSuffix(input[:cursorPos], " ") {
		return m.completeCommand(tokens[0])
	}

	// Check for git-specific completion
	if len(tokens) >= 1 && tokens[0] == "git" {
		return m.completeGit(tokens, cursorPos, input)
	}

	// Otherwise, complete files/directories
	var prefix string
	if len(tokens) > 0 {
		prefix = tokens[len(tokens)-1]
	}

	return m.completeFile(prefix)
}

// completeCommand provides command completions
func (m *Manager) completeCommand(prefix string) ([]string, error) {
	var completions []string

	// Add built-in commands
	builtins := []string{
		"cd", "pwd", "exit", "help", "history", "alias", "export",
	}

	for _, builtin := range builtins {
		if strings.HasPrefix(builtin, prefix) {
			completions = append(completions, builtin)
		}
	}

	// Add aliases
	for alias := range m.config.Aliases {
		if strings.HasPrefix(alias, prefix) {
			completions = append(completions, alias)
		}
	}

	// Add commands from PATH
	pathCompletions, err := m.completeFromPath(prefix)
	if err == nil {
		completions = append(completions, pathCompletions...)
	}

	// Remove duplicates and sort
	completions = m.removeDuplicates(completions)
	sort.Strings(completions)

	return completions, nil
}

// completeFromPath finds executable commands in PATH
func (m *Manager) completeFromPath(prefix string) ([]string, error) {
	var completions []string
	seen := make(map[string]bool)

	for _, dir := range m.config.PathDirs {
		if dir == "" {
			continue
		}

		entries, err := os.ReadDir(dir)
		if err != nil {
			continue // Skip directories we can't read
		}

		for _, entry := range entries {
			name := entry.Name()

			// Skip if doesn't match prefix
			if !strings.HasPrefix(name, prefix) {
				continue
			}

			// Skip if already seen
			if seen[name] {
				continue
			}

			// Check if it's executable
			if m.isExecutable(filepath.Join(dir, name), entry) {
				completions = append(completions, name)
				seen[name] = true
			}
		}
	}

	return completions, nil
}

// isExecutable checks if a file is executable
func (m *Manager) isExecutable(_ string, entry fs.DirEntry) bool {
	if entry.IsDir() {
		return false
	}

	info, err := entry.Info()
	if err != nil {
		return false
	}

	mode := info.Mode()
	return mode&0111 != 0 // Check if any execute bit is set
}

// completeFile provides file and directory completions
func (m *Manager) completeFile(prefix string) ([]string, error) {
	var completions []string

	// Determine the directory and filename prefix
	dir := "."
	filePrefix := prefix

	if strings.Contains(prefix, "/") {
		dir = filepath.Dir(prefix)
		filePrefix = filepath.Base(prefix)
	}

	// Expand ~ to home directory
	if strings.HasPrefix(dir, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		dir = filepath.Join(homeDir, dir[2:])
	} else if dir == "~" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		dir = homeDir
	}

	// Read directory entries
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		name := entry.Name()

		// Skip hidden files unless configured to show them
		if !m.config.CompletionShowHidden && strings.HasPrefix(name, ".") {
			continue
		}

		// Check if name matches prefix
		matches := false
		if m.config.CompletionCaseInsensitive {
			matches = strings.HasPrefix(strings.ToLower(name), strings.ToLower(filePrefix))
		} else {
			matches = strings.HasPrefix(name, filePrefix)
		}

		if !matches {
			continue
		}

		// Build the full completion
		var completion string
		if dir == "." {
			completion = name
		} else if strings.HasPrefix(prefix, "/") || strings.HasPrefix(prefix, "~/") {
			completion = filepath.Join(dir, name)
		} else {
			completion = filepath.Join(dir, name)
		}

		// Add trailing slash for directories
		if entry.IsDir() {
			completion += "/"
		}

		completions = append(completions, completion)
	}

	sort.Strings(completions)
	return completions, nil
}

// removeDuplicates removes duplicate strings from a slice
func (m *Manager) removeDuplicates(slice []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}

// GetCommonPrefix returns the common prefix of all completions
func (m *Manager) GetCommonPrefix(completions []string) string {
	if len(completions) == 0 {
		return ""
	}

	if len(completions) == 1 {
		return completions[0]
	}

	// Find the common prefix
	prefix := completions[0]
	for _, completion := range completions[1:] {
		prefix = m.commonPrefix(prefix, completion)
		if prefix == "" {
			break
		}
	}

	return prefix
}

// commonPrefix finds the common prefix between two strings
func (m *Manager) commonPrefix(a, b string) string {
	minLen := len(a)
	if len(b) < minLen {
		minLen = len(b)
	}

	for i := 0; i < minLen; i++ {
		if a[i] != b[i] {
			return a[:i]
		}
	}

	return a[:minLen]
}

// FormatCompletions formats completions for display
func (m *Manager) FormatCompletions(completions []string, maxWidth int) []string {
	if len(completions) == 0 {
		return nil
	}

	// If only one completion, return it as-is
	if len(completions) == 1 {
		return completions
	}

	// Calculate column width
	maxLen := 0
	for _, completion := range completions {
		if len(completion) > maxLen {
			maxLen = len(completion)
		}
	}

	// Add some padding
	colWidth := maxLen + 2

	// Calculate number of columns
	cols := maxWidth / colWidth
	if cols < 1 {
		cols = 1
	}

	// Format into columns
	var formatted []string
	var currentLine strings.Builder

	for i, completion := range completions {
		if i > 0 && i%cols == 0 {
			formatted = append(formatted, currentLine.String())
			currentLine.Reset()
		}

		currentLine.WriteString(fmt.Sprintf("%-*s", colWidth, completion))
	}

	if currentLine.Len() > 0 {
		formatted = append(formatted, currentLine.String())
	}

	return formatted
}

// completeGit provides git-specific completions
func (m *Manager) completeGit(tokens []string, cursorPos int, input string) ([]string, error) {
	if len(tokens) < 2 {
		// Complete git subcommands
		return m.completeGitSubcommands("")
	}

	subcommand := tokens[1]

	// If we're still completing the subcommand
	if len(tokens) == 2 && !strings.HasSuffix(input[:cursorPos], " ") {
		return m.completeGitSubcommands(subcommand)
	}

	// Complete based on git subcommand
	switch subcommand {
	case "checkout", "co", "switch":
		return m.completeGitBranches(tokens)
	case "branch":
		return m.completeGitBranches(tokens)
	case "merge":
		return m.completeGitBranches(tokens)
	case "add":
		return m.completeGitModifiedFiles(tokens)
	case "commit":
		return m.completeGitCommitOptions(tokens)
	case "push", "pull":
		return m.completeGitRemotes(tokens)
	case "remote":
		return m.completeGitRemoteSubcommands(tokens)
	case "log", "show", "diff":
		return m.completeGitRefs(tokens)
	default:
		// Default to file completion for other git commands
		var prefix string
		if len(tokens) > 0 {
			prefix = tokens[len(tokens)-1]
		}
		return m.completeFile(prefix)
	}
}

// completeGitSubcommands completes git subcommands
func (m *Manager) completeGitSubcommands(prefix string) ([]string, error) {
	subcommands := []string{
		"add", "branch", "checkout", "clone", "commit", "diff", "fetch",
		"init", "log", "merge", "pull", "push", "rebase", "remote",
		"reset", "show", "status", "switch", "tag",
	}

	var completions []string
	for _, cmd := range subcommands {
		if strings.HasPrefix(cmd, prefix) {
			completions = append(completions, cmd)
		}
	}

	return completions, nil
}

// completeGitBranches completes git branch names
func (m *Manager) completeGitBranches(tokens []string) ([]string, error) {
	// Use git manager if available
	if m.config.GitEnabled {
		// This would integrate with the git manager
		// For now, return common branch names
		branches := []string{"main", "master", "develop", "feature/", "bugfix/", "hotfix/"}

		var prefix string
		if len(tokens) > 2 {
			prefix = tokens[len(tokens)-1]
		}

		var completions []string
		for _, branch := range branches {
			if strings.HasPrefix(branch, prefix) {
				completions = append(completions, branch)
			}
		}
		return completions, nil
	}
	return nil, nil
}

// completeGitModifiedFiles completes modified files for git add
func (m *Manager) completeGitModifiedFiles(tokens []string) ([]string, error) {
	// For now, fall back to regular file completion
	// This could be enhanced to only show modified files
	var prefix string
	if len(tokens) > 2 {
		prefix = tokens[len(tokens)-1]
	}
	return m.completeFile(prefix)
}

// completeGitCommitOptions completes git commit options
func (m *Manager) completeGitCommitOptions(tokens []string) ([]string, error) {
	options := []string{"-m", "--message", "-a", "--all", "--amend", "-v", "--verbose"}

	var prefix string
	if len(tokens) > 2 {
		prefix = tokens[len(tokens)-1]
	}

	var completions []string
	for _, option := range options {
		if strings.HasPrefix(option, prefix) {
			completions = append(completions, option)
		}
	}
	return completions, nil
}

// completeGitRemotes completes git remote names
func (m *Manager) completeGitRemotes(tokens []string) ([]string, error) {
	remotes := []string{"origin", "upstream"}

	var prefix string
	if len(tokens) > 2 {
		prefix = tokens[len(tokens)-1]
	}

	var completions []string
	for _, remote := range remotes {
		if strings.HasPrefix(remote, prefix) {
			completions = append(completions, remote)
		}
	}
	return completions, nil
}

// completeGitRemoteSubcommands completes git remote subcommands
func (m *Manager) completeGitRemoteSubcommands(tokens []string) ([]string, error) {
	subcommands := []string{"add", "remove", "rename", "show", "prune", "update"}

	var prefix string
	if len(tokens) > 2 {
		prefix = tokens[len(tokens)-1]
	}

	var completions []string
	for _, cmd := range subcommands {
		if strings.HasPrefix(cmd, prefix) {
			completions = append(completions, cmd)
		}
	}
	return completions, nil
}

// completeGitRefs completes git references (branches, tags, commits)
func (m *Manager) completeGitRefs(tokens []string) ([]string, error) {
	// Combine branches and common refs
	refs := []string{"HEAD", "main", "master", "develop", "origin/main", "origin/master"}

	var prefix string
	if len(tokens) > 2 {
		prefix = tokens[len(tokens)-1]
	}

	var completions []string
	for _, ref := range refs {
		if strings.HasPrefix(ref, prefix) {
			completions = append(completions, ref)
		}
	}
	return completions, nil
}
