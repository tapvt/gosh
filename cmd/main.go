// Package main provides the entry point for the gosh shell.
// Gosh is a modern, feature-rich shell written in Go that combines
// the familiarity of bash with modern conveniences and performance.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gosh/internal/config"
	"gosh/internal/shell"
)

const (
	// Version represents the current version of gosh
	Version = "1.0.0"

	// DefaultConfigDir is the default directory for gosh configuration files
	DefaultConfigDir = ".config/gosh"

	// DefaultDirPermissions is the default permission for created directories
	DefaultDirPermissions = 0750
)

var (
	// Command line flags
	versionFlag = flag.Bool("version", false, "Show version information")
	configFlag  = flag.String("config", "", "Path to configuration file")
	debugFlag   = flag.Bool("debug", false, "Enable debug mode")
	helpFlag    = flag.Bool("help", false, "Show help information")
)

func main() {
	flag.Parse()

	// Handle version flag
	if *versionFlag {
		fmt.Printf("gosh version %s\n", Version)
		fmt.Println("A modern shell written in Go")
		os.Exit(0)
	}

	// Handle help flag
	if *helpFlag {
		showHelp()
		os.Exit(0)
	}

	// Initialize configuration
	cfg, err := initializeConfig()
	if err != nil {
		log.Fatalf("Failed to initialize configuration: %v", err)
	}

	// Set debug mode if requested
	if *debugFlag {
		cfg.Debug = true
	}

	// Create and start the shell
	sh, err := shell.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create shell: %v", err)
	}

	// Run the shell
	if err := sh.Run(); err != nil {
		log.Fatalf("Shell execution failed: %v", err)
	}
}

// initializeConfig sets up the configuration for gosh
func initializeConfig() (*config.Config, error) {
	var configPath string

	// Use provided config path or find default
	if *configFlag != "" {
		configPath = *configFlag
	} else {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		configPath = filepath.Join(homeDir, DefaultConfigDir)
	}

	// Load configuration
	cfg, err := config.Load(configPath)
	if err != nil {
		// If config doesn't exist, create default
		if os.IsNotExist(err) {
			cfg = config.Default()
			cfg.ConfigDir = configPath

			// Create config directory if it doesn't exist
			if mkdirErr := os.MkdirAll(configPath, DefaultDirPermissions); mkdirErr != nil {
				return nil, fmt.Errorf("failed to create config directory: %w", mkdirErr)
			}
		} else {
			return nil, fmt.Errorf("failed to load configuration: %w", err)
		}
	}

	return cfg, nil
}

// showHelp displays help information for gosh
func showHelp() {
	fmt.Printf("gosh - A modern shell written in Go (version %s)\n\n", Version)
	fmt.Println("Usage:")
	fmt.Println("  gosh [options]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -version     Show version information")
	fmt.Println("  -config      Path to configuration file")
	fmt.Println("  -debug       Enable debug mode")
	fmt.Println("  -help        Show this help message")
	fmt.Println()
	fmt.Println("Configuration:")
	fmt.Println("  Gosh looks for configuration files in the following order:")
	fmt.Println("  1. ~/.config/gosh/goshrc")
	fmt.Println("  2. ~/.goshrc")
	fmt.Println("  3. ~/.gosh_profile (login shells)")
	fmt.Println()
	fmt.Println("Built-in Commands:")
	fmt.Println("  cd, pwd, exit, help, history, alias, export")
	fmt.Println()
	fmt.Println("Features:")
	fmt.Println("  - Tab completion for commands and files")
	fmt.Println("  - Git integration and status display")
	fmt.Println("  - Customizable prompts")
	fmt.Println("  - Command history with search")
	fmt.Println("  - Bash-compatible configuration files")
	fmt.Println()
	fmt.Println("For more information, visit: https://github.com/yourusername/gosh")
}
