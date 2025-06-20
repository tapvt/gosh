// Package config provides configuration management for gosh.
// It handles loading and parsing of configuration files like .goshrc and .gosh_profile,
// as well as managing runtime configuration options.
package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	// KeyValueParts is the expected number of parts when splitting key=value pairs
	KeyValueParts = 2
)

// Config holds all configuration options for gosh
type Config struct {
	// Core settings
	ConfigDir   string `json:"config_dir"`
	Debug       bool   `json:"debug"`
	ShowWelcome bool   `json:"show_welcome"`

	// Prompt settings
	PromptFormat  string `json:"prompt_format"`
	ShowGitInfo   bool   `json:"show_git_info"`
	ShowTimestamp bool   `json:"show_timestamp"`
	PromptColor   string `json:"prompt_color"`

	// History settings
	HistorySize       int    `json:"history_size"`
	HistoryFile       string `json:"history_file"`
	SaveHistory       bool   `json:"save_history"`
	HistoryDuplicates bool   `json:"history_duplicates"`

	// Completion settings
	CompletionEnabled         bool `json:"completion_enabled"`
	CompletionCaseInsensitive bool `json:"completion_case_insensitive"`
	CompletionShowHidden      bool `json:"completion_show_hidden"`

	// Git integration settings
	GitEnabled    bool `json:"git_enabled"`
	GitShowStatus bool `json:"git_show_status"`
	GitShowBranch bool `json:"git_show_branch"`
	GitShowAhead  bool `json:"git_show_ahead"`

	// Environment variables
	Environment map[string]string `json:"environment"`

	// Aliases
	Aliases map[string]string `json:"aliases"`

	// Path settings
	PathDirs []string `json:"path_dirs"`
}

// Default returns a default configuration
func Default() *Config {
	homeDir, _ := os.UserHomeDir()

	return &Config{
		// Core settings
		ConfigDir:   filepath.Join(homeDir, ".config", "gosh"),
		Debug:       false,
		ShowWelcome: true,

		// Prompt settings
		PromptFormat:  "%u@%h:%w%g$ ",
		ShowGitInfo:   true,
		ShowTimestamp: false,
		PromptColor:   "auto",

		// History settings
		HistorySize:       10000,
		HistoryFile:       filepath.Join(homeDir, ".gosh_history"),
		SaveHistory:       true,
		HistoryDuplicates: false,

		// Completion settings
		CompletionEnabled:         true,
		CompletionCaseInsensitive: true,
		CompletionShowHidden:      false,

		// Git integration settings
		GitEnabled:    true,
		GitShowStatus: true,
		GitShowBranch: true,
		GitShowAhead:  true,

		// Environment variables
		Environment: make(map[string]string),

		// Aliases
		Aliases: map[string]string{
			"ll": "ls -la",
			"la": "ls -A",
			"l":  "ls -CF",
		},

		// Path settings
		PathDirs: strings.Split(os.Getenv("PATH"), ":"),
	}
}

// Load loads configuration from the specified directory
func Load(configDir string) (*Config, error) {
	cfg := Default()
	cfg.ConfigDir = configDir

	// Try to load from various config file locations
	configFiles := []string{
		filepath.Join(configDir, "config"),
		filepath.Join(configDir, "goshrc"),
	}

	// Also check home directory for .goshrc
	if homeDir, err := os.UserHomeDir(); err == nil {
		configFiles = append(configFiles, filepath.Join(homeDir, ".goshrc"))
	}

	var loaded bool
	for _, configFile := range configFiles {
		if err := cfg.loadFromFile(configFile); err == nil {
			loaded = true
			break
		}
	}

	if !loaded {
		return nil, os.ErrNotExist
	}

	return cfg, nil
}

// loadFromFile loads configuration from a specific file
func (c *Config) loadFromFile(filename string) error {
	// Validate the file path to prevent directory traversal
	cleanPath := filepath.Clean(filename)
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("invalid file path: %s", filename)
	}

	file, err := os.Open(cleanPath)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if err := c.parseLine(line); err != nil {
			return fmt.Errorf("error parsing line %d: %w", lineNum, err)
		}
	}

	return scanner.Err()
}

// parseLine parses a single configuration line
func (c *Config) parseLine(line string) error {
	// Handle export statements
	if strings.HasPrefix(line, "export ") {
		return c.parseExport(strings.TrimPrefix(line, "export "))
	}

	// Handle alias statements
	if strings.HasPrefix(line, "alias ") {
		return c.parseAlias(strings.TrimPrefix(line, "alias "))
	}

	// Handle set statements for gosh-specific settings
	if strings.HasPrefix(line, "set ") {
		return c.parseSet(strings.TrimPrefix(line, "set "))
	}

	// Handle direct variable assignments
	if strings.Contains(line, "=") {
		return c.parseAssignment(line)
	}

	return nil
}

// parseExport parses export statements
func (c *Config) parseExport(line string) error {
	parts := strings.SplitN(line, "=", KeyValueParts)
	if len(parts) != KeyValueParts {
		return fmt.Errorf("invalid export statement: %s", line)
	}

	key := strings.TrimSpace(parts[0])
	value := strings.Trim(strings.TrimSpace(parts[1]), "\"'")

	c.Environment[key] = value
	return nil
}

// parseAlias parses alias statements
func (c *Config) parseAlias(line string) error {
	parts := strings.SplitN(line, "=", KeyValueParts)
	if len(parts) != KeyValueParts {
		return fmt.Errorf("invalid alias statement: %s", line)
	}

	key := strings.TrimSpace(parts[0])
	value := strings.Trim(strings.TrimSpace(parts[1]), "\"'")

	c.Aliases[key] = value
	return nil
}

// parseSet parses set statements for gosh-specific settings
func (c *Config) parseSet(line string) error {
	parts := strings.SplitN(line, "=", KeyValueParts)
	if len(parts) != KeyValueParts {
		return fmt.Errorf("invalid set statement: %s", line)
	}

	key := strings.TrimSpace(parts[0])
	value := strings.Trim(strings.TrimSpace(parts[1]), "\"'")

	return c.setConfigValue(key, value)
}

// parseAssignment parses direct variable assignments
func (c *Config) parseAssignment(line string) error {
	parts := strings.SplitN(line, "=", KeyValueParts)
	if len(parts) != KeyValueParts {
		return fmt.Errorf("invalid assignment: %s", line)
	}

	key := strings.TrimSpace(parts[0])
	value := strings.Trim(strings.TrimSpace(parts[1]), "\"'")

	// Check if it's a gosh-specific setting
	if strings.HasPrefix(key, "GOSH_") {
		return c.setConfigValue(strings.TrimPrefix(key, "GOSH_"), value)
	}

	// Otherwise, treat as environment variable
	c.Environment[key] = value
	return nil
}

// setConfigValue sets a configuration value by key
func (c *Config) setConfigValue(key, value string) error {
	upperKey := strings.ToUpper(key)

	// Handle core settings
	if err := c.setCoreSettings(upperKey, value); err == nil {
		return nil
	}

	// Handle prompt settings
	if err := c.setPromptSettings(upperKey, value); err == nil {
		return nil
	}

	// Handle history settings
	if err := c.setHistorySettings(upperKey, value); err == nil {
		return nil
	}

	// Handle completion settings
	if err := c.setCompletionSettings(upperKey, value); err == nil {
		return nil
	}

	// Handle git settings
	if err := c.setGitSettings(upperKey, value); err == nil {
		return nil
	}

	return fmt.Errorf("unknown configuration key: %s", key)
}

// setCoreSettings handles core configuration settings
func (c *Config) setCoreSettings(key, value string) error {
	switch key {
	case "DEBUG":
		c.Debug = parseBool(value)
		return nil
	case "SHOW_WELCOME":
		c.ShowWelcome = parseBool(value)
		return nil
	default:
		return fmt.Errorf("not a core setting")
	}
}

// setPromptSettings handles prompt configuration settings
func (c *Config) setPromptSettings(key, value string) error {
	switch key {
	case "PROMPT_FORMAT":
		c.PromptFormat = value
		return nil
	case "SHOW_GIT_INFO":
		c.ShowGitInfo = parseBool(value)
		return nil
	case "SHOW_TIMESTAMP":
		c.ShowTimestamp = parseBool(value)
		return nil
	case "PROMPT_COLOR":
		c.PromptColor = value
		return nil
	default:
		return fmt.Errorf("not a prompt setting")
	}
}

// setHistorySettings handles history configuration settings
func (c *Config) setHistorySettings(key, value string) error {
	switch key {
	case "HISTORY_SIZE":
		if size, err := strconv.Atoi(value); err == nil {
			c.HistorySize = size
		}
		return nil
	case "HISTORY_FILE":
		c.HistoryFile = value
		return nil
	case "SAVE_HISTORY":
		c.SaveHistory = parseBool(value)
		return nil
	case "HISTORY_DUPLICATES":
		c.HistoryDuplicates = parseBool(value)
		return nil
	default:
		return fmt.Errorf("not a history setting")
	}
}

// setCompletionSettings handles completion configuration settings
func (c *Config) setCompletionSettings(key, value string) error {
	switch key {
	case "COMPLETION_ENABLED":
		c.CompletionEnabled = parseBool(value)
		return nil
	case "COMPLETION_CASE_INSENSITIVE":
		c.CompletionCaseInsensitive = parseBool(value)
		return nil
	case "COMPLETION_SHOW_HIDDEN":
		c.CompletionShowHidden = parseBool(value)
		return nil
	default:
		return fmt.Errorf("not a completion setting")
	}
}

// setGitSettings handles git configuration settings
func (c *Config) setGitSettings(key, value string) error {
	switch key {
	case "GIT_ENABLED":
		c.GitEnabled = parseBool(value)
		return nil
	case "GIT_SHOW_STATUS":
		c.GitShowStatus = parseBool(value)
		return nil
	case "GIT_SHOW_BRANCH":
		c.GitShowBranch = parseBool(value)
		return nil
	case "GIT_SHOW_AHEAD":
		c.GitShowAhead = parseBool(value)
		return nil
	default:
		return fmt.Errorf("not a git setting")
	}
}

// parseBool parses a boolean value from string
func parseBool(value string) bool {
	switch strings.ToLower(value) {
	case "true", "yes", "1", "on":
		return true
	default:
		return false
	}
}
