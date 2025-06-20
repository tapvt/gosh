# Gosh - A Modern Shell Written in Go

Gosh is a feature-rich, modern shell implementation written in Go that combines the familiarity of bash with modern conveniences and performance.

## Features

- 🚀 **Fast and Lightweight**: Built in Go for excellent performance
- 🔧 **Bash-Compatible**: Supports familiar bash syntax and features
- 📝 **Smart Tab Completion**: Intelligent completion for commands, files, directories, and git operations
- 🎨 **Customizable Prompts**: Rich prompt system with git integration and color schemes
- 📁 **Configuration Files**: Support for `.goshrc` and `.gosh_profile`
- 🔀 **Git Integration**: Built-in git status, branch information, and git-aware completion
- 📚 **Command History**: Persistent history with search capabilities and arrow key navigation
- 🛠️ **Built-in Commands**: Essential commands like cd, pwd, exit, help, history, alias, export
- ⚡ **Interactive Features**: Proper Ctrl+C handling, variable expansion ($VAR), and error recovery
- 🔍 **Smart Error Handling**: Helpful error messages with command suggestions

## Quick Start

### Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/gosh.git
cd gosh

# Build and install
make install
```

### Usage

```bash
# Start gosh
gosh

# Or run directly from repository (for development)
make run

# Or run the setup script for first-time configuration
./setup.sh
```

### Interactive Features

Gosh provides a rich interactive experience:

- **Tab Completion**: Press Tab to complete commands, files, and git operations
  - `git che<Tab>` → `git checkout`
  - `cd /ho<Tab>` → `cd /home/`
  - `h<Tab>` → shows `help` and `history`

- **Command History**: Use arrow keys to navigate through command history
  - `↑` / `↓` to browse previous/next commands
  - Persistent history across sessions

- **Variable Expansion**: Use environment variables in commands
  - `echo $HOME` → shows your home directory
  - `export MY_VAR=value` then `echo $MY_VAR` → shows `value`

- **Ctrl+C Handling**: Properly interrupts commands and returns to prompt

- **Smart Error Messages**: Get helpful suggestions when commands fail
  - `xit` → "Did you mean: exit?"

## Configuration

Gosh supports two configuration files:

- **`.gosh_profile`**: Executed once at login (like `.bash_profile`)
- **`.goshrc`**: Executed for each interactive shell (like `.bashrc`)

### Example Configuration

```bash
# ~/.goshrc
export GOSH_PROMPT_FORMAT="%u@%h:%w%g$ "
export GOSH_HISTORY_SIZE=10000
alias ll="ls -la"
alias grep="grep --color=auto"
```

## Built-in Commands

- `cd` - Change directory
- `pwd` - Print working directory
- `exit` - Exit the shell
- `help` - Show help information
- `history` - Show command history
- `alias` - Create command aliases
- `export` - Set environment variables

## Git Integration

Gosh provides seamless git integration:

- Current branch displayed in prompt
- Git status indicators (clean, dirty, ahead/behind)
- Tab completion for git commands and branches
- Git-aware directory navigation

## Development

### Building

```bash
make build        # Build the binary
make test         # Run tests
make clean        # Clean build artifacts
make install      # Install to system
```

### Project Structure

```
gosh/
├── cmd/           # Command-line interface
├── internal/      # Internal packages
│   ├── shell/     # Core shell logic
│   ├── parser/    # Command parsing
│   ├── completion/# Tab completion
│   ├── prompt/    # Prompt system
│   ├── config/    # Configuration management
│   ├── git/       # Git integration
│   └── history/   # History management
├── docs/          # Documentation
├── scripts/       # Setup and utility scripts
└── Makefile       # Build system
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

MIT License - see LICENSE file for details.

## Acknowledgments

Inspired by bash, zsh, and fish shells, built with modern Go practices.
