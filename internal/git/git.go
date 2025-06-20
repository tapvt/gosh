// Package git provides git integration functionality for gosh.
// It handles git repository detection, status checking, and branch information.
package git

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"gosh/internal/config"
)

const (
	// MinStatusLineLength is the minimum length for a git status line
	MinStatusLineLength = 2
	// ExpectedRevListParts is the expected number of parts from git rev-list output
	ExpectedRevListParts = 2
)

// Info represents git repository information
type Info struct {
	Branch         string
	HasUncommitted bool
	HasUntracked   bool
	HasStaged      bool
	Ahead          int
	Behind         int
	IsRepo         bool
}

// Manager handles git operations and information gathering
type Manager struct {
	config *config.Config
}

// New creates a new git manager
func New(cfg *config.Config) (*Manager, error) {
	return &Manager{
		config: cfg,
	}, nil
}

// GetInfo returns git information for the current directory
func (m *Manager) GetInfo() (*Info, error) {
	if !m.config.GitEnabled {
		return nil, nil
	}

	// Check if we're in a git repository
	if !m.isGitRepo() {
		return nil, nil
	}

	info := &Info{IsRepo: true}

	// Get branch name
	branch, err := m.getCurrentBranch()
	if err == nil {
		info.Branch = branch
	}

	// Get status information
	if err := m.getStatus(info); err != nil && m.config.Debug {
		fmt.Fprintf(os.Stderr, "Warning: failed to get git status: %v\n", err)
	}

	// Get ahead/behind information
	if err := m.getAheadBehind(info); err != nil && m.config.Debug {
		fmt.Fprintf(os.Stderr, "Warning: failed to get ahead/behind info: %v\n", err)
	}

	return info, nil
}

// isGitRepo checks if the current directory is in a git repository
func (m *Manager) isGitRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Stderr = nil // Suppress error output
	return cmd.Run() == nil
}

// getCurrentBranch returns the current git branch name
func (m *Manager) getCurrentBranch() (string, error) {
	cmd := exec.Command("git", "symbolic-ref", "--short", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		// Try to get commit hash if not on a branch
		cmd = exec.Command("git", "rev-parse", "--short", "HEAD")
		output, err = cmd.Output()
		if err != nil {
			return "", err
		}
		return "(" + strings.TrimSpace(string(output)) + ")", nil
	}
	return strings.TrimSpace(string(output)), nil
}

// getStatus gets the git status information
func (m *Manager) getStatus(info *Info) error {
	cmd := exec.Command("git", "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if len(line) < MinStatusLineLength {
			continue
		}

		staged := line[0]
		unstaged := line[1]

		// Check for staged changes
		if staged != ' ' && staged != '?' {
			info.HasStaged = true
		}

		// Check for unstaged changes
		if unstaged != ' ' && unstaged != '?' {
			info.HasUncommitted = true
		}

		// Check for untracked files
		if staged == '?' && unstaged == '?' {
			info.HasUntracked = true
		}
	}

	return nil
}

// getAheadBehind gets ahead/behind information relative to upstream
func (m *Manager) getAheadBehind(info *Info) error {
	cmd := exec.Command("git", "rev-list", "--count", "--left-right", "@{upstream}...HEAD")
	output, err := cmd.Output()
	if err != nil {
		// No upstream configured, that's okay
		return nil
	}

	parts := strings.Fields(strings.TrimSpace(string(output)))
	if len(parts) != ExpectedRevListParts {
		return fmt.Errorf("unexpected git rev-list output: %s", output)
	}

	behind, err := strconv.Atoi(parts[0])
	if err != nil {
		return fmt.Errorf("failed to parse behind count: %w", err)
	}

	ahead, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("failed to parse ahead count: %w", err)
	}

	info.Behind = behind
	info.Ahead = ahead

	return nil
}

// GetBranches returns a list of git branches for completion
func (m *Manager) GetBranches() ([]string, error) {
	return m.getGitCommandOutput("git", "branch", "--format=%(refname:short)")
}

// GetRemotes returns a list of git remotes for completion
func (m *Manager) GetRemotes() ([]string, error) {
	if !m.isGitRepo() {
		return nil, nil
	}

	cmd := exec.Command("git", "remote")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var remotes []string
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		remote := strings.TrimSpace(scanner.Text())
		if remote != "" {
			remotes = append(remotes, remote)
		}
	}

	return remotes, scanner.Err()
}

// GetModifiedFiles returns a list of modified files for completion
func (m *Manager) GetModifiedFiles() ([]string, error) {
	return m.getGitCommandOutput("git", "diff", "--name-only")
}

// getGitCommandOutput executes a git command and returns the output as a slice of strings
func (m *Manager) getGitCommandOutput(name string, args ...string) ([]string, error) {
	if !m.isGitRepo() {
		return nil, nil
	}

	cmd := exec.Command(name, args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var results []string
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			results = append(results, line)
		}
	}

	return results, scanner.Err()
}

// GetUntrackedFiles returns a list of untracked files for completion
func (m *Manager) GetUntrackedFiles() ([]string, error) {
	if !m.isGitRepo() {
		return nil, nil
	}

	cmd := exec.Command("git", "ls-files", "--others", "--exclude-standard")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var files []string
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		file := strings.TrimSpace(scanner.Text())
		if file != "" {
			files = append(files, file)
		}
	}

	return files, scanner.Err()
}

// FindGitRoot finds the root directory of the git repository
func (m *Manager) FindGitRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// IsIgnored checks if a file is ignored by git
func (m *Manager) IsIgnored(path string) (bool, error) {
	if !m.isGitRepo() {
		return false, nil
	}

	cmd := exec.Command("git", "check-ignore", path)
	err := cmd.Run()
	if err != nil {
		// If the command fails, the file is not ignored
		return false, nil
	}
	return true, nil
}

// GetRepoInfo returns general repository information
func (m *Manager) GetRepoInfo() (map[string]string, error) {
	if !m.isGitRepo() {
		return nil, fmt.Errorf("not a git repository")
	}

	info := make(map[string]string)

	// Get repository root
	if root, err := m.FindGitRoot(); err == nil {
		info["root"] = root
		info["name"] = filepath.Base(root)
	}

	// Get current branch
	if branch, err := m.getCurrentBranch(); err == nil {
		info["branch"] = branch
	}

	// Get remote URL
	cmd := exec.Command("git", "remote", "get-url", "origin")
	if output, err := cmd.Output(); err == nil {
		info["origin"] = strings.TrimSpace(string(output))
	}

	// Get last commit
	cmd = exec.Command("git", "log", "-1", "--pretty=format:%h %s")
	if output, err := cmd.Output(); err == nil {
		info["last_commit"] = strings.TrimSpace(string(output))
	}

	return info, nil
}
