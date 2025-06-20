// Package history provides command history management for gosh.
// It handles persistent storage, search functionality, and history navigation.
package history

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gosh/internal/config"
)

const (
	// HistoryLineParts is the expected number of parts in a history line
	HistoryLineParts = 3
	// DefaultFilePermissions is the default permission for created files
	DefaultFilePermissions = 0600
	// DefaultDirPermissions is the default permission for created directories
	DefaultDirPermissions = 0750
)

// Entry represents a single history entry
type Entry struct {
	Command   string
	Timestamp time.Time
	Directory string
}

// GetCommand returns the command string (implements parser.HistoryEntry)
func (e Entry) GetCommand() string {
	return e.Command
}

// GetTimestamp returns the timestamp as a string (implements parser.HistoryEntry)
func (e Entry) GetTimestamp() string {
	return e.Timestamp.Format(time.RFC3339)
}

// Manager handles command history operations
type Manager struct {
	config  *config.Config
	entries []Entry
	current int // Current position in history for navigation
}

// New creates a new history manager
func New(cfg *config.Config) (*Manager, error) {
	mgr := &Manager{
		config:  cfg,
		entries: make([]Entry, 0, cfg.HistorySize),
		current: -1,
	}

	// Load existing history if configured
	if cfg.SaveHistory {
		if err := mgr.load(); err != nil && cfg.Debug {
			fmt.Fprintf(os.Stderr, "Warning: failed to load history: %v\n", err)
		}
	}

	return mgr, nil
}

// Add adds a command to the history
func (m *Manager) Add(command string) {
	command = strings.TrimSpace(command)
	if command == "" {
		return
	}

	// Skip duplicates if configured
	if !m.config.HistoryDuplicates && len(m.entries) > 0 {
		if m.entries[len(m.entries)-1].Command == command {
			return
		}
	}

	// Get current directory
	wd, _ := os.Getwd()

	entry := Entry{
		Command:   command,
		Timestamp: time.Now(),
		Directory: wd,
	}

	// Add to entries
	m.entries = append(m.entries, entry)

	// Trim if exceeding max size
	if len(m.entries) > m.config.HistorySize {
		m.entries = m.entries[len(m.entries)-m.config.HistorySize:]
	}

	// Reset current position
	m.current = len(m.entries)

	// Save to file if configured
	if m.config.SaveHistory {
		if err := m.save(); err != nil && m.config.Debug {
			fmt.Fprintf(os.Stderr, "Warning: failed to save history: %v\n", err)
		}
	}
}

// GetAll returns all history entries
func (m *Manager) GetAll() []Entry {
	return m.entries
}

// GetRecent returns the most recent n entries
func (m *Manager) GetRecent(n int) []Entry {
	if n <= 0 || len(m.entries) == 0 {
		return nil
	}

	start := len(m.entries) - n
	if start < 0 {
		start = 0
	}

	return m.entries[start:]
}

// Search searches for commands containing the given term
func (m *Manager) Search(term string) []Entry {
	if term == "" {
		return nil
	}

	var matches []Entry
	term = strings.ToLower(term)

	for _, entry := range m.entries {
		if strings.Contains(strings.ToLower(entry.Command), term) {
			matches = append(matches, entry)
		}
	}

	return matches
}

// SearchPrefix searches for commands starting with the given prefix
func (m *Manager) SearchPrefix(prefix string) []Entry {
	if prefix == "" {
		return nil
	}

	var matches []Entry
	prefix = strings.ToLower(prefix)

	for _, entry := range m.entries {
		if strings.HasPrefix(strings.ToLower(entry.Command), prefix) {
			matches = append(matches, entry)
		}
	}

	return matches
}

// Previous returns the previous command in history
func (m *Manager) Previous() string {
	if len(m.entries) == 0 {
		return ""
	}

	if m.current > 0 {
		m.current--
	}

	if m.current >= 0 && m.current < len(m.entries) {
		return m.entries[m.current].Command
	}

	return ""
}

// Next returns the next command in history
func (m *Manager) Next() string {
	if len(m.entries) == 0 {
		return ""
	}

	if m.current < len(m.entries)-1 {
		m.current++
		return m.entries[m.current].Command
	}

	// Reset to end of history
	m.current = len(m.entries)
	return ""
}

// Reset resets the history navigation position
func (m *Manager) Reset() {
	m.current = len(m.entries)
}

// Clear clears all history
func (m *Manager) Clear() error {
	m.entries = make([]Entry, 0, m.config.HistorySize)
	m.current = -1

	// Clear the history file if it exists
	if m.config.SaveHistory {
		return m.clearFile()
	}

	return nil
}

// load loads history from the configured file
func (m *Manager) load() error {
	if m.config.HistoryFile == "" {
		return nil
	}

	file, err := os.Open(m.config.HistoryFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist yet, that's okay
		}
		return err
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Parse the line format: timestamp|directory|command
		parts := strings.SplitN(line, "|", HistoryLineParts)
		if len(parts) < HistoryLineParts {
			// Old format, just the command
			entry := Entry{
				Command:   line,
				Timestamp: time.Now(),
				Directory: "",
			}
			m.entries = append(m.entries, entry)
			continue
		}

		// Parse timestamp
		timestamp, err := time.Parse(time.RFC3339, parts[0])
		if err != nil {
			timestamp = time.Now()
		}

		entry := Entry{
			Command:   parts[2],
			Timestamp: timestamp,
			Directory: parts[1],
		}

		m.entries = append(m.entries, entry)
	}

	// Trim if exceeding max size
	if len(m.entries) > m.config.HistorySize {
		m.entries = m.entries[len(m.entries)-m.config.HistorySize:]
	}

	m.current = len(m.entries)
	return scanner.Err()
}

// save saves history to the configured file
func (m *Manager) save() error {
	if m.config.HistoryFile == "" {
		return nil
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(m.config.HistoryFile)
	if err := os.MkdirAll(dir, DefaultDirPermissions); err != nil {
		return err
	}

	file, err := os.Create(m.config.HistoryFile)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	for _, entry := range m.entries {
		line := fmt.Sprintf("%s|%s|%s\n",
			entry.Timestamp.Format(time.RFC3339),
			entry.Directory,
			entry.Command)
		if _, err := file.WriteString(line); err != nil {
			return err
		}
	}

	return nil
}

// clearFile clears the history file
func (m *Manager) clearFile() error {
	if m.config.HistoryFile == "" {
		return nil
	}

	return os.Remove(m.config.HistoryFile)
}

// GetStats returns history statistics
func (m *Manager) GetStats() map[string]interface{} {
	stats := make(map[string]interface{})

	stats["total_entries"] = len(m.entries)
	stats["max_size"] = m.config.HistorySize
	stats["save_enabled"] = m.config.SaveHistory
	stats["duplicates_allowed"] = m.config.HistoryDuplicates

	if len(m.entries) > 0 {
		stats["oldest_entry"] = m.entries[0].Timestamp
		stats["newest_entry"] = m.entries[len(m.entries)-1].Timestamp
	}

	// Count unique commands
	unique := make(map[string]bool)
	for _, entry := range m.entries {
		unique[entry.Command] = true
	}
	stats["unique_commands"] = len(unique)

	return stats
}

// Export exports history to a file in a specific format
func (m *Manager) Export(filename, format string) error {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, DefaultFilePermissions)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	switch format {
	case "bash":
		// Export in bash history format
		for _, entry := range m.entries {
			if _, err := fmt.Fprintf(file, "%s\n", entry.Command); err != nil {
				return err
			}
		}
	case "json":
		// Export in JSON format (simplified)
		if _, err := file.WriteString("[\n"); err != nil {
			return err
		}
		for i, entry := range m.entries {
			line := fmt.Sprintf(`  {"command": %q, "timestamp": %q, "directory": %q}`,
				strings.ReplaceAll(entry.Command, `"`, `\"`),
				entry.Timestamp.Format(time.RFC3339),
				entry.Directory)
			if i < len(m.entries)-1 {
				line += ","
			}
			line += "\n"
			if _, err := file.WriteString(line); err != nil {
				return err
			}
		}
		if _, err := file.WriteString("]\n"); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported export format: %s", format)
	}

	return nil
}
