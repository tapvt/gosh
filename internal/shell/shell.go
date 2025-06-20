// Package shell provides the core shell functionality for gosh.
// It implements the main shell loop, command execution, and integration
// with other shell components like completion, history, and prompts.
package shell

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"gosh/internal/completion"
	"gosh/internal/config"
	"gosh/internal/history"
	"gosh/internal/parser"
	"gosh/internal/prompt"

	"github.com/chzyer/readline"
)

// shellCompleter implements readline.AutoCompleter for tab completion
type shellCompleter struct {
	completion *completion.Manager
}

// Do implements the AutoCompleter interface
func (c *shellCompleter) Do(line []rune, pos int) (newLine [][]rune, length int) {
	lineStr := string(line)
	completions, err := c.completion.Complete(lineStr, pos)
	if err != nil || len(completions) == 0 {
		return nil, 0
	}

	// Convert completions to [][]rune
	var result [][]rune
	for _, completion := range completions {
		result = append(result, []rune(completion))
	}

	// Find common prefix length
	if len(completions) > 0 {
		commonPrefix := c.completion.GetCommonPrefix(completions)
		// Calculate how much of the current word to replace
		// Find the start of the current word being completed
		wordStart := pos
		for wordStart > 0 && lineStr[wordStart-1] != ' ' && lineStr[wordStart-1] != '\t' {
			wordStart--
		}

		// Get the current partial word
		currentWord := lineStr[wordStart:pos]

		// Only set length if the common prefix actually starts with the current word
		if strings.HasPrefix(commonPrefix, currentWord) {
			length = len(currentWord)
		}
	}

	return result, length
}

// Shell represents the main shell instance
type Shell struct {
	config     *config.Config
	history    *history.Manager
	prompt     *prompt.Manager
	completion *completion.Manager
	parser     *parser.Parser
	readline   *readline.Instance
	writer     io.Writer
	ctx        context.Context
	cancel     context.CancelFunc
}

// New creates a new shell instance with the given configuration
func New(cfg *config.Config) (*Shell, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// Initialize history manager
	historyMgr, err := history.New(cfg)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize history: %w", err)
	}

	// Initialize prompt manager
	promptMgr, err := prompt.New(cfg)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize prompt: %w", err)
	}

	// Initialize completion manager
	completionMgr, err := completion.New(cfg)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to initialize completion: %w", err)
	}

	// Initialize parser
	parserInst := parser.New(cfg)

	// Create readline instance with completion
	rl, err := readline.NewEx(&readline.Config{
		Prompt:          "> ",
		HistoryFile:     cfg.HistoryFile,
		AutoComplete:    &shellCompleter{completion: completionMgr},
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create readline: %w", err)
	}

	shell := &Shell{
		config:     cfg,
		history:    historyMgr,
		prompt:     promptMgr,
		completion: completionMgr,
		parser:     parserInst,
		readline:   rl,
		writer:     os.Stdout,
		ctx:        ctx,
		cancel:     cancel,
	}

	return shell, nil
}

// Run starts the main shell loop
func (s *Shell) Run() error {
	defer s.cancel()
	defer s.readline.Close()

	// Setup signal handling
	s.setupSignalHandling()

	// Load configuration files
	if err := s.loadConfigFiles(); err != nil {
		return fmt.Errorf("failed to load config files: %w", err)
	}

	// Print welcome message if configured
	if s.config.ShowWelcome {
		s.printWelcome()
	}

	// Main shell loop
	return s.mainLoop()
}

// mainLoop implements the main read-eval-print loop
func (s *Shell) mainLoop() error {
	for {
		select {
		case <-s.ctx.Done():
			return nil
		default:
			// Read input (readline handles prompt generation)
			input, err := s.readInput()
			if err != nil {
				if err == readline.ErrInterrupt {
					// Handle Ctrl+C
					continue
				}
				if err == io.EOF {
					fmt.Fprintln(s.writer, "\nGoodbye!")
					return nil
				}
				return fmt.Errorf("failed to read input: %w", err)
			}

			// Skip empty input
			if strings.TrimSpace(input) == "" {
				continue
			}

			// Add to history
			s.history.Add(input)

			// Parse and execute command
			if err := s.executeCommand(input); err != nil {
				// Enhanced error handling with context
				s.handleError(err, input)
			}
		}
	}
}

// readInput reads a line of input from the user
func (s *Shell) readInput() (string, error) {
	// Update prompt before reading
	promptStr, err := s.prompt.Generate()
	if err != nil {
		promptStr = "gosh> "
	}
	s.readline.SetPrompt(promptStr)

	line, err := s.readline.Readline()
	if err != nil {
		return "", err
	}
	return line, nil
}

// executeCommand parses and executes a command
func (s *Shell) executeCommand(input string) error {
	// Parse the command
	cmd, err := s.parser.Parse(input)
	if err != nil {
		return fmt.Errorf("parse error: %w", err)
	}

	// Execute the command
	return cmd.Execute(s.ctx, s.config)
}

// setupSignalHandling sets up signal handlers for the shell
func (s *Shell) setupSignalHandling() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for {
			select {
			case sig := <-sigChan:
				switch sig {
				case syscall.SIGINT:
					// Handle Ctrl+C gracefully
					fmt.Fprintln(s.writer, "^C")
					// Don't exit, just interrupt current operation and continue
					// The main loop will continue and show a new prompt
				case syscall.SIGTERM:
					// Handle termination
					fmt.Fprintln(s.writer, "\nTerminating gosh...")
					s.cancel()
					return
				}
			case <-s.ctx.Done():
				return
			}
		}
	}()
}

// loadConfigFiles loads .goshrc and .gosh_profile files
func (s *Shell) loadConfigFiles() error {
	// Load .gosh_profile (login shell)
	if err := s.loadConfigFile(".gosh_profile"); err != nil && s.config.Debug {
		fmt.Fprintf(s.writer, "Warning: failed to load .gosh_profile: %v\n", err)
	}

	// Load .goshrc (interactive shell)
	if err := s.loadConfigFile(".goshrc"); err != nil && s.config.Debug {
		fmt.Fprintf(s.writer, "Warning: failed to load .goshrc: %v\n", err)
	}

	return nil
}

// loadConfigFile loads a specific configuration file
func (s *Shell) loadConfigFile(filename string) error {
	// Implementation will be added when we implement the config system
	// For now, just return nil
	return nil
}

// printWelcome prints the welcome message
func (s *Shell) printWelcome() {
	fmt.Fprintf(s.writer, "Welcome to gosh - A modern shell written in Go\n")
	fmt.Fprintf(s.writer, "Type 'help' for available commands.\n\n")
}

// handleError provides enhanced error handling with context and recovery
func (s *Shell) handleError(err error, input string) {
	// Categorize and handle different types of errors
	errorMsg := err.Error()

	switch {
	case strings.Contains(errorMsg, "command not found"):
		fmt.Fprintf(s.writer, "gosh: %s\n", errorMsg)
		if s.config.Debug {
			fmt.Fprintf(s.writer, "Debug: Input was '%s'\n", input)
		}
		s.suggestSimilarCommands(input)

	case strings.Contains(errorMsg, "no such file or directory"):
		fmt.Fprintf(s.writer, "gosh: %s\n", errorMsg)
		if s.config.Debug {
			fmt.Fprintf(s.writer, "Debug: Check if the path exists and you have permission\n")
		}

	case strings.Contains(errorMsg, "permission denied"):
		fmt.Fprintf(s.writer, "gosh: %s\n", errorMsg)
		if s.config.Debug {
			fmt.Fprintf(s.writer, "Debug: Check file permissions or try with sudo\n")
		}

	case strings.Contains(errorMsg, "parse error"):
		fmt.Fprintf(s.writer, "gosh: syntax error: %s\n", errorMsg)
		if s.config.Debug {
			fmt.Fprintf(s.writer, "Debug: Check your command syntax\n")
		}

	default:
		// Generic error handling
		fmt.Fprintf(s.writer, "gosh: %s\n", errorMsg)
		if s.config.Debug {
			fmt.Fprintf(s.writer, "Debug: Error type: %T\n", err)
		}
	}
}

// suggestSimilarCommands suggests similar commands when a command is not found
func (s *Shell) suggestSimilarCommands(input string) {
	tokens := strings.Fields(input)
	if len(tokens) == 0 {
		return
	}

	command := tokens[0]
	suggestions := []string{}

	// Check built-in commands for similarity
	builtins := []string{"cd", "pwd", "exit", "help", "history", "alias", "export"}
	for _, builtin := range builtins {
		if s.isSimilar(command, builtin) {
			suggestions = append(suggestions, builtin)
		}
	}

	// Check aliases
	for alias := range s.config.Aliases {
		if s.isSimilar(command, alias) {
			suggestions = append(suggestions, alias)
		}
	}

	if len(suggestions) > 0 {
		fmt.Fprintf(s.writer, "Did you mean: %s?\n", strings.Join(suggestions, ", "))
	}
}

// isSimilar checks if two strings are similar (simple Levenshtein-like check)
func (s *Shell) isSimilar(a, b string) bool {
	if len(a) == 0 || len(b) == 0 {
		return false
	}

	// Simple similarity check: same first character and similar length
	if a[0] == b[0] && abs(len(a)-len(b)) <= 2 {
		return true
	}

	// Check for common prefixes
	minLen := len(a)
	if len(b) < minLen {
		minLen = len(b)
	}

	if minLen >= 2 {
		return a[:2] == b[:2]
	}

	return false
}

// abs returns the absolute value of an integer
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
