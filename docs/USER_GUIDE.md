# Gosh User Guide

Welcome to gosh! This guide will help you get the most out of your new shell.

## Getting Started

### Starting Gosh

```bash
# Start gosh
gosh

# Start with debug mode
gosh --debug

# Show version
gosh --version

# Show help
gosh --help
```

### Basic Usage

Gosh works like any other shell. You can run commands, navigate directories, and use pipes and redirections:

```bash
# Basic commands
ls -la
cd /home/user
pwd
echo "Hello, World!"

# Pipes and redirections
ls -la | grep "txt"
echo "content" > file.txt
cat file.txt
```

## Built-in Commands

### Core Commands

- **`cd [directory]`**: Change directory
  ```bash
  cd /home/user
  cd ..
  cd ~
  cd  # Goes to home directory
  ```

- **`pwd`**: Print working directory
  ```bash
  pwd
  ```

- **`exit`**: Exit the shell
  ```bash
  exit
  ```

- **`help`**: Show help information
  ```bash
  help
  ```

### History Commands

- **`history`**: Show command history
  ```bash
  history          # Show all history
  history 10       # Show last 10 commands
  history | grep git  # Search history
  ```

### Alias Management

- **`alias`**: Manage command aliases
  ```bash
  alias                    # Show all aliases
  alias ll="ls -la"       # Create alias
  alias gs="git status"   # Git alias
  ```

### Environment Variables

- **`export`**: Set environment variables
  ```bash
  export PATH="/new/path:$PATH"
  export EDITOR=vim
  export GOSH_PROMPT_FORMAT="%u@%h:%w$ "
  ```

## Tab Completion

Gosh provides intelligent tab completion for:

### Command Completion
```bash
gi<Tab>     # Completes to git (if available)
ls<Tab>     # Shows ls options
```

### File and Directory Completion
```bash
cd /ho<Tab>     # Completes to /home/
ls *.t<Tab>     # Shows all .txt, .tar files, etc.
```

### Git-Aware Completion
```bash
git checkout <Tab>    # Shows available branches
git add <Tab>         # Shows modified files
```

## Prompt Customization

### Prompt Format Codes

- `%u` - Username
- `%h` - Hostname
- `%w` - Full working directory path
- `%W` - Working directory basename
- `%g` - Git information
- `%t` - Timestamp
- `%$` - Prompt character ($ or # for root)

### Examples

```bash
# Simple prompt
export GOSH_PROMPT_FORMAT="%u$ "

# Full information prompt
export GOSH_PROMPT_FORMAT="%u@%h:%w%g [%t]$ "

# Minimal prompt
export GOSH_PROMPT_FORMAT="%W%g$ "
```

### Color Schemes

```bash
# Auto coloring (default)
export GOSH_PROMPT_COLOR=auto

# Minimal coloring
export GOSH_PROMPT_COLOR=minimal

# Bright colors
export GOSH_PROMPT_COLOR=bright

# No colors
export GOSH_PROMPT_COLOR=none
```

## Git Integration

### Git Information in Prompt

When in a git repository, your prompt can show:

- Current branch name
- Repository status indicators:
  - `*` - Uncommitted changes
  - `?` - Untracked files
  - `+` - Staged changes
- Ahead/behind indicators:
  - `↑3` - 3 commits ahead of remote
  - `↓2` - 2 commits behind remote

### Example Git Prompts

```bash
user@host:~/project (main *)$           # Dirty working directory
user@host:~/project (main ↑2)$          # 2 commits ahead
user@host:~/project (feature-branch)$   # On feature branch
user@host:~/project (abc1234)$          # Detached HEAD
```

### Git Configuration

```bash
# Enable git integration
export GOSH_GIT_ENABLED=true

# Show git status in prompt
export GOSH_GIT_SHOW_STATUS=true

# Show current branch
export GOSH_GIT_SHOW_BRANCH=true

# Show ahead/behind information
export GOSH_GIT_SHOW_AHEAD=true
```

## History Management

### History Configuration

```bash
# Set history size
export GOSH_HISTORY_SIZE=10000

# Set history file location
export GOSH_HISTORY_FILE=~/.gosh_history

# Enable history saving
export GOSH_SAVE_HISTORY=true

# Allow duplicate entries
export GOSH_HISTORY_DUPLICATES=false
```

### History Navigation

- **Up Arrow**: Previous command
- **Down Arrow**: Next command
- **Ctrl+R**: Search history (if implemented)

### History Commands

```bash
# Show recent history
history 20

# Search history
history | grep "git"

# Clear history
history -c
```

## Configuration Files

### .goshrc

Interactive shell configuration (loaded for each shell session):

```bash
# ~/.goshrc
export GOSH_PROMPT_FORMAT="%u@%h:%w%g$ "
alias ll="ls -la"
alias gs="git status"

# Custom functions
mkcd() {
    mkdir -p "$1" && cd "$1"
}
```

### .gosh_profile

Login shell configuration (loaded once at login):

```bash
# ~/.gosh_profile
export PATH="/usr/local/bin:$PATH"
export EDITOR=vim
export GOPATH="$HOME/go"

# Load .goshrc for interactive shells
if [ -f ~/.goshrc ]; then
    source ~/.goshrc
fi
```

## Advanced Features

### Custom Functions

Define functions in your `.goshrc`:

```bash
# Create directory and cd into it
mkcd() {
    mkdir -p "$1" && cd "$1"
}

# Find files by name
ff() {
    find . -name "*$1*" -type f
}

# Git commit with message
gcm() {
    git commit -m "$1"
}
```

### Conditional Configuration

```bash
# Load different configs based on hostname
case $(hostname) in
    work-laptop)
        source ~/.goshrc.work
        ;;
    home-desktop)
        source ~/.goshrc.home
        ;;
esac

# Load local customizations
if [ -f ~/.goshrc.local ]; then
    source ~/.goshrc.local
fi
```

### Environment-Specific Settings

```bash
# Development environment
if [ -d "$HOME/dev" ]; then
    export DEVPATH="$HOME/dev"
    alias dev="cd $DEVPATH"
fi

# Work-specific settings
if [ "$USER" = "work-user" ]; then
    export WORK_PROXY="http://proxy.company.com:8080"
    export HTTP_PROXY="$WORK_PROXY"
    export HTTPS_PROXY="$WORK_PROXY"
fi
```

## Tips and Tricks

### Productivity Tips

1. **Use aliases for common commands:**
   ```bash
   alias g="git"
   alias k="kubectl"
   alias d="docker"
   ```

2. **Create project shortcuts:**
   ```bash
   alias proj="cd ~/projects/current-project"
   alias logs="tail -f /var/log/app.log"
   ```

3. **Use functions for complex operations:**
   ```bash
   backup() {
       tar -czf "backup-$(date +%Y%m%d).tar.gz" "$1"
   }
   ```

### Git Workflow Integration

```bash
# Quick git aliases
alias gs="git status"
alias ga="git add"
alias gc="git commit"
alias gp="git push"
alias gl="git log --oneline"

# Advanced git functions
gac() {
    git add . && git commit -m "$1"
}

gacp() {
    git add . && git commit -m "$1" && git push
}
```

### Directory Navigation

```bash
# Quick navigation
alias ..="cd .."
alias ...="cd ../.."
alias ....="cd ../../.."

# Bookmark directories
alias proj="cd ~/projects"
alias docs="cd ~/documents"
alias down="cd ~/downloads"
```

## Troubleshooting

### Common Issues

1. **Prompt not showing git info:**
   - Check if you're in a git repository
   - Verify `GOSH_GIT_ENABLED=true`
   - Ensure git is installed

2. **Tab completion not working:**
   - Check `GOSH_COMPLETION_ENABLED=true`
   - Verify file permissions
   - Try in a different directory

3. **Configuration not loading:**
   - Check file paths and permissions
   - Use `gosh --debug` for diagnostics
   - Verify syntax in config files

### Debug Mode

Enable debug mode to see what's happening:

```bash
gosh --debug
```

Or set in configuration:
```bash
export GOSH_DEBUG=true
```

## Getting Help

- Use `help` command for built-in help
- Check the documentation in the `docs/` directory
- Open issues on GitHub for bugs or feature requests
- Join community discussions for tips and tricks
