// Package prompt provides customizable prompt functionality for gosh.
// It supports various prompt formats, git integration, and color customization.
package prompt

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"gosh/internal/config"
	"gosh/internal/git"
)

const (
	// PromptColorNone represents no color formatting
	PromptColorNone = "none"
	// UnknownValue represents unknown values in prompts
	UnknownValue = "unknown"
)

// Manager handles prompt generation and customization
type Manager struct {
	config     *config.Config
	gitManager *git.Manager
}

// New creates a new prompt manager
func New(cfg *config.Config) (*Manager, error) {
	gitMgr, err := git.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize git manager: %w", err)
	}

	return &Manager{
		config:     cfg,
		gitManager: gitMgr,
	}, nil
}

// Generate generates the current prompt string
func (m *Manager) Generate() (string, error) {
	format := m.getPromptFormat()
	prompt := m.processPromptFormat(format)

	// Apply colors if enabled
	if m.config.PromptColor != PromptColorNone {
		prompt = m.applyColors(prompt)
	}

	return prompt, nil
}

// getPromptFormat returns the prompt format, using default if empty
func (m *Manager) getPromptFormat() string {
	format := m.config.PromptFormat
	if format == "" {
		format = "%u@%h:%w%g$ "
	}
	return format
}

// processPromptFormat processes the prompt format string and expands escape sequences
func (m *Manager) processPromptFormat(format string) string {
	var result strings.Builder
	i := 0
	for i < len(format) {
		if format[i] == '%' && i+1 < len(format) {
			expansion := m.expandEscapeSequence(format[i+1])
			result.WriteString(expansion)
			i += 2
		} else {
			result.WriteByte(format[i])
			i++
		}
	}
	return result.String()
}

// expandEscapeSequence expands a single escape sequence character
func (m *Manager) expandEscapeSequence(char byte) string {
	switch char {
	case 'u':
		return m.getUsername()
	case 'h':
		return m.getHostname()
	case 'w':
		return m.getWorkingDir()
	case 'W':
		return m.getWorkingDirBasename()
	case 'g':
		return m.getGitInfoSafe()
	case 't':
		return m.getTimestampSafe()
	case '$':
		return m.getPromptChar()
	case '%':
		return "%"
	default:
		// Unknown escape, just write it as-is
		return "%" + string(char)
	}
}

// getGitInfoSafe returns git info if enabled, empty string otherwise
func (m *Manager) getGitInfoSafe() string {
	if !m.config.ShowGitInfo {
		return ""
	}
	gitInfo, err := m.getGitInfo()
	if err != nil || gitInfo == "" {
		return ""
	}
	return gitInfo
}

// getTimestampSafe returns timestamp if enabled, empty string otherwise
func (m *Manager) getTimestampSafe() string {
	if !m.config.ShowTimestamp {
		return ""
	}
	return m.getTimestamp()
}

// getUsername returns the current username
func (m *Manager) getUsername() string {
	if currentUser, err := user.Current(); err == nil {
		return currentUser.Username
	}
	return UnknownValue
}

// getHostname returns the hostname
func (m *Manager) getHostname() string {
	if hostname, err := os.Hostname(); err == nil {
		return hostname
	}
	return "localhost"
}

// getWorkingDir returns the current working directory
func (m *Manager) getWorkingDir() string {
	wd, err := os.Getwd()
	if err != nil {
		return UnknownValue
	}

	// Replace home directory with ~
	if homeDir, err := os.UserHomeDir(); err == nil {
		if strings.HasPrefix(wd, homeDir) {
			wd = "~" + wd[len(homeDir):]
		}
	}

	return wd
}

// getWorkingDirBasename returns just the basename of the working directory
func (m *Manager) getWorkingDirBasename() string {
	wd, err := os.Getwd()
	if err != nil {
		return UnknownValue
	}

	// Special case for home directory
	if homeDir, err := os.UserHomeDir(); err == nil && wd == homeDir {
		return "~"
	}

	return filepath.Base(wd)
}

// getGitInfo returns git status information
func (m *Manager) getGitInfo() (string, error) {
	if !m.config.GitEnabled {
		return "", nil
	}

	info, err := m.gitManager.GetInfo()
	if err != nil {
		return "", err
	}

	if info == nil {
		return "", nil
	}

	var parts []string

	// Add branch name
	if m.config.GitShowBranch && info.Branch != "" {
		parts = append(parts, info.Branch)
	}

	// Add status indicators
	if m.config.GitShowStatus {
		var indicators []string

		if info.HasUncommitted {
			indicators = append(indicators, "*")
		}

		if info.HasUntracked {
			indicators = append(indicators, "?")
		}

		if info.HasStaged {
			indicators = append(indicators, "+")
		}

		if len(indicators) > 0 {
			parts = append(parts, strings.Join(indicators, ""))
		}
	}

	// Add ahead/behind info
	if m.config.GitShowAhead {
		if info.Ahead > 0 {
			parts = append(parts, fmt.Sprintf("↑%d", info.Ahead))
		}
		if info.Behind > 0 {
			parts = append(parts, fmt.Sprintf("↓%d", info.Behind))
		}
	}

	if len(parts) == 0 {
		return "", nil
	}

	return " (" + strings.Join(parts, " ") + ")", nil
}

// getTimestamp returns the current timestamp
func (m *Manager) getTimestamp() string {
	return time.Now().Format("15:04:05")
}

// getPromptChar returns the appropriate prompt character
func (m *Manager) getPromptChar() string {
	// Use $ for regular users, # for root
	if os.Geteuid() == 0 {
		return "#"
	}
	return "$"
}

// applyColors applies color formatting to the prompt
func (m *Manager) applyColors(prompt string) string {
	if m.config.PromptColor == PromptColorNone || m.config.PromptColor == "off" {
		return prompt
	}

	// Define color codes
	colors := map[string]string{
		"reset":   "\033[0m",
		"bold":    "\033[1m",
		"red":     "\033[31m",
		"green":   "\033[32m",
		"yellow":  "\033[33m",
		"blue":    "\033[34m",
		"magenta": "\033[35m",
		"cyan":    "\033[36m",
		"white":   "\033[37m",
	}

	// Apply default coloring scheme
	switch m.config.PromptColor {
	case "auto", "default":
		// Color username@hostname in green
		prompt = strings.Replace(prompt, m.getUsername()+"@"+m.getHostname(),
			colors["green"]+m.getUsername()+"@"+m.getHostname()+colors["reset"], 1)

		// Color working directory in blue
		wd := m.getWorkingDir()
		prompt = strings.Replace(prompt, wd, colors["blue"]+wd+colors["reset"], 1)

		// Color git info in yellow
		if strings.Contains(prompt, "(") && strings.Contains(prompt, ")") {
			start := strings.Index(prompt, "(")
			end := strings.LastIndex(prompt, ")") + 1
			if start < end {
				gitPart := prompt[start:end]
				coloredGit := colors["yellow"] + gitPart + colors["reset"]
				prompt = prompt[:start] + coloredGit + prompt[end:]
			}
		}

	case "minimal":
		// Just color the prompt character
		prompt = strings.Replace(prompt, "$", colors["green"]+"$"+colors["reset"], -1)
		prompt = strings.Replace(prompt, "#", colors["red"]+"#"+colors["reset"], -1)

	case "bright":
		// Use bright colors
		prompt = colors["bold"] + prompt + colors["reset"]
	}

	return prompt
}

// SetFormat updates the prompt format
func (m *Manager) SetFormat(format string) {
	m.config.PromptFormat = format
}

// GetAvailableFormats returns available prompt format options
func (m *Manager) GetAvailableFormats() map[string]string {
	return map[string]string{
		"%u": "Username",
		"%h": "Hostname",
		"%w": "Full working directory path",
		"%W": "Working directory basename",
		"%g": "Git information",
		"%t": "Timestamp (HH:MM:SS)",
		"%$": "Prompt character ($ or # for root)",
		"%%": "Literal % character",
	}
}
