// Package parser provides command parsing functionality for gosh.
// It handles parsing of command lines, including pipes, redirections,
// and built-in command recognition.
package parser

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"gosh/internal/config"
)

const (
	// KeyValueParts is the expected number of parts when splitting key=value pairs
	KeyValueParts = 2
)

// Command represents a parsed command
type Command interface {
	Execute(ctx context.Context, cfg *config.Config) error
}

// Parser handles parsing of command lines
type Parser struct {
	config *config.Config
}

// New creates a new parser instance
func New(cfg *config.Config) *Parser {
	return &Parser{
		config: cfg,
	}
}

// Parse parses a command line and returns a Command
func (p *Parser) Parse(input string) (Command, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return &NoOpCommand{}, nil
	}

	// Handle aliases
	if expanded, ok := p.expandAlias(input); ok {
		input = expanded
	}

	// Split into tokens
	tokens, err := p.tokenize(input)
	if err != nil {
		return nil, err
	}

	if len(tokens) == 0 {
		return &NoOpCommand{}, nil
	}

	// Check for built-in commands
	if builtin := p.parseBuiltin(tokens); builtin != nil {
		return builtin, nil
	}

	// Parse as external command
	return p.parseExternal(tokens)
}

// tokenize splits input into tokens, handling quotes and escapes
func (p *Parser) tokenize(input string) ([]string, error) {
	var tokens []string
	var current strings.Builder
	var inQuotes bool
	var quoteChar rune
	var escaped bool

	for _, r := range input {
		if escaped {
			current.WriteRune(r)
			escaped = false
			continue
		}

		if r == '\\' {
			escaped = true
			continue
		}

		if !inQuotes && (r == '"' || r == '\'') {
			inQuotes = true
			quoteChar = r
			continue
		}

		if inQuotes && r == quoteChar {
			inQuotes = false
			continue
		}

		if !inQuotes && (r == ' ' || r == '\t') {
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
			continue
		}

		current.WriteRune(r)
	}

	if inQuotes {
		return nil, fmt.Errorf("unclosed quote")
	}

	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}

	return tokens, nil
}

// expandAlias expands aliases if present
func (p *Parser) expandAlias(input string) (string, bool) {
	tokens := strings.Fields(input)
	if len(tokens) == 0 {
		return input, false
	}

	if expansion, ok := p.config.Aliases[tokens[0]]; ok {
		tokens[0] = expansion
		return strings.Join(tokens, " "), true
	}

	return input, false
}

// parseBuiltin checks if the command is a built-in and returns it
func (p *Parser) parseBuiltin(tokens []string) Command {
	cmd := tokens[0]
	args := tokens[1:]

	switch cmd {
	case "cd":
		return &CdCommand{Args: args}
	case "pwd":
		return &PwdCommand{}
	case "exit":
		return &ExitCommand{Args: args}
	case "help":
		return &HelpCommand{Args: args}
	case "history":
		return &HistoryCommand{Args: args}
	case "alias":
		return &AliasCommand{Args: args, Config: p.config}
	case "export":
		return &ExportCommand{Args: args, Config: p.config}
	default:
		return nil
	}
}

// parseExternal parses an external command
func (p *Parser) parseExternal(tokens []string) (Command, error) {
	// Expand variables in tokens
	expandedTokens := make([]string, len(tokens))
	for i, token := range tokens {
		expandedTokens[i] = p.expandVariables(token)
	}

	return &ExternalCommand{
		Name: expandedTokens[0],
		Args: expandedTokens[1:],
	}, nil
}

// expandVariables expands environment variables in a token
func (p *Parser) expandVariables(token string) string {
	// Simple variable expansion for $VAR and ${VAR}
	result := token

	// Handle $VAR format
	for i := 0; i < len(result); i++ {
		if result[i] == '$' && i+1 < len(result) {
			if result[i+1] == '{' {
				// Handle ${VAR} format
				end := strings.Index(result[i+2:], "}")
				if end != -1 {
					varName := result[i+2 : i+2+end]
					varValue := p.getVariable(varName)
					result = result[:i] + varValue + result[i+3+end:]
					i += len(varValue) - 1
				}
			} else {
				// Handle $VAR format
				start := i + 1
				end := start
				for end < len(result) && (isAlphaNumeric(result[end]) || result[end] == '_') {
					end++
				}
				if end > start {
					varName := result[start:end]
					varValue := p.getVariable(varName)
					result = result[:i] + varValue + result[end:]
					i += len(varValue) - 1
				}
			}
		}
	}

	return result
}

// getVariable gets a variable value from environment or config
func (p *Parser) getVariable(name string) string {
	// First check config environment
	if value, ok := p.config.Environment[name]; ok {
		return value
	}

	// Then check system environment
	return os.Getenv(name)
}

// isAlphaNumeric checks if a character is alphanumeric
func isAlphaNumeric(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}

// NoOpCommand represents a no-operation command
type NoOpCommand struct{}

// Execute implements the Command interface for NoOpCommand
func (c *NoOpCommand) Execute(_ context.Context, _ *config.Config) error {
	return nil
}

// CdCommand implements the cd built-in command
type CdCommand struct {
	Args []string
}

// Execute implements the Command interface for CdCommand
func (c *CdCommand) Execute(_ context.Context, _ *config.Config) error {
	var dir string
	if len(c.Args) == 0 {
		// No arguments, go to home directory
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		dir = homeDir
	} else {
		dir = c.Args[0]
	}

	// Expand ~ to home directory
	if strings.HasPrefix(dir, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		dir = filepath.Join(homeDir, dir[2:])
	}

	if err := os.Chdir(dir); err != nil {
		return fmt.Errorf("cd: %w", err)
	}

	return nil
}

// PwdCommand implements the pwd built-in command
type PwdCommand struct{}

// Execute implements the Command interface for PwdCommand
func (c *PwdCommand) Execute(_ context.Context, _ *config.Config) error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("pwd: %w", err)
	}
	fmt.Println(wd)
	return nil
}

// ExitCommand implements the exit built-in command
type ExitCommand struct {
	Args []string
}

// Execute implements the Command interface for ExitCommand
func (c *ExitCommand) Execute(_ context.Context, _ *config.Config) error {
	os.Exit(0)
	return nil
}

// HelpCommand implements the help built-in command
type HelpCommand struct {
	Args []string
}

// Execute implements the Command interface for HelpCommand
func (c *HelpCommand) Execute(_ context.Context, _ *config.Config) error {
	fmt.Println("Gosh - A modern shell written in Go")
	fmt.Println()
	fmt.Println("Built-in commands:")
	fmt.Println("  cd [dir]     Change directory")
	fmt.Println("  pwd          Print working directory")
	fmt.Println("  exit         Exit the shell")
	fmt.Println("  help         Show this help message")
	fmt.Println("  history      Show command history")
	fmt.Println("  alias        Manage command aliases")
	fmt.Println("  export       Set environment variables")
	fmt.Println()
	fmt.Println("Features:")
	fmt.Println("  - Tab completion (press Tab)")
	fmt.Println("  - Command history (use arrow keys)")
	fmt.Println("  - Git integration in prompt")
	fmt.Println("  - Customizable configuration")
	return nil
}

// HistoryCommand implements the history built-in command
type HistoryCommand struct {
	Args    []string
	Manager HistoryManager // Interface to history manager
}

// HistoryManager interface for accessing history
type HistoryManager interface {
	GetAll() []HistoryEntry
	GetRecent(n int) []HistoryEntry
	Search(term string) []HistoryEntry
	Clear() error
}

// HistoryEntry represents a history entry
type HistoryEntry interface {
	GetCommand() string
	GetTimestamp() string
}

// Execute implements the Command interface for HistoryCommand
func (c *HistoryCommand) Execute(_ context.Context, _ *config.Config) error {
	if c.Manager == nil {
		fmt.Println("History functionality not available")
		return nil
	}

	if len(c.Args) == 0 {
		// Show all history
		entries := c.Manager.GetAll()
		for i, entry := range entries {
			fmt.Printf("%4d  %s\n", i+1, entry.GetCommand())
		}
		return nil
	}

	// Handle history subcommands
	switch c.Args[0] {
	case "-c", "clear":
		return c.Manager.Clear()
	default:
		// Try to parse as number for recent entries
		if n, err := strconv.Atoi(c.Args[0]); err == nil {
			entries := c.Manager.GetRecent(n)
			for i, entry := range entries {
				fmt.Printf("%4d  %s\n", len(c.Manager.GetAll())-len(entries)+i+1, entry.GetCommand())
			}
			return nil
		}

		// Search for term
		entries := c.Manager.Search(c.Args[0])
		for _, entry := range entries {
			fmt.Printf("  %s\n", entry.GetCommand())
		}
	}

	return nil
}

// AliasCommand implements the alias built-in command
type AliasCommand struct {
	Args   []string
	Config *config.Config
}

// Execute implements the Command interface for AliasCommand
func (c *AliasCommand) Execute(_ context.Context, _ *config.Config) error {
	if len(c.Args) == 0 {
		// Show all aliases
		for name, value := range c.Config.Aliases {
			fmt.Printf("alias %s='%s'\n", name, value)
		}
		return nil
	}

	// Parse alias definition
	arg := strings.Join(c.Args, " ")
	parts := strings.SplitN(arg, "=", KeyValueParts)
	if len(parts) != KeyValueParts {
		return fmt.Errorf("alias: invalid format, use: alias name=value")
	}

	name := strings.TrimSpace(parts[0])
	value := strings.Trim(strings.TrimSpace(parts[1]), "\"'")

	c.Config.Aliases[name] = value
	return nil
}

// ExportCommand implements the export built-in command
type ExportCommand struct {
	Args   []string
	Config *config.Config
}

// Execute implements the Command interface for ExportCommand
func (c *ExportCommand) Execute(_ context.Context, _ *config.Config) error {
	if len(c.Args) == 0 {
		// Show all environment variables
		for key, value := range c.Config.Environment {
			fmt.Printf("export %s='%s'\n", key, value)
		}
		return nil
	}

	// Parse export definition
	arg := strings.Join(c.Args, " ")
	parts := strings.SplitN(arg, "=", KeyValueParts)
	if len(parts) != KeyValueParts {
		return fmt.Errorf("export: invalid format, use: export NAME=value")
	}

	name := strings.TrimSpace(parts[0])
	value := strings.Trim(strings.TrimSpace(parts[1]), "\"'")

	c.Config.Environment[name] = value
	if err := os.Setenv(name, value); err != nil {
		return fmt.Errorf("failed to set environment variable %s: %w", name, err)
	}
	return nil
}

// ExternalCommand represents an external command
type ExternalCommand struct {
	Name string
	Args []string
}

// Execute implements the Command interface for ExternalCommand
func (c *ExternalCommand) Execute(ctx context.Context, _ *config.Config) error {
	cmd := exec.CommandContext(ctx, c.Name, c.Args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		// Provide more user-friendly error messages
		if exitError, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("command '%s' exited with code %d", c.Name, exitError.ExitCode())
		}
		if err.Error() == "exec: \""+c.Name+"\": executable file not found in $PATH" {
			return fmt.Errorf("command not found: %s", c.Name)
		}
		return fmt.Errorf("failed to execute '%s': %w", c.Name, err)
	}
	return nil
}
