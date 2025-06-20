# Gosh Installation Guide

This guide covers various methods to install and set up gosh on your system.

## Prerequisites

- Go 1.19 or later
- Git (for git integration features)
- A Unix-like operating system (Linux, macOS, BSD)

## Installation Methods

### Method 1: Build from Source (Recommended)

1. **Clone the repository:**
   ```bash
   git clone https://github.com/yourusername/gosh.git
   cd gosh
   ```

2. **Build and install:**
   ```bash
   make install
   ```

3. **Verify installation:**
   ```bash
   gosh --version
   ```

### Method 2: Manual Build

1. **Clone and build:**
   ```bash
   git clone https://github.com/yourusername/gosh.git
   cd gosh
   go build -o gosh cmd/main.go
   ```

2. **Install manually:**
   ```bash
   sudo cp gosh /usr/local/bin/
   sudo chmod +x /usr/local/bin/gosh
   ```

### Method 3: Go Install

```bash
go install github.com/yourusername/gosh@latest
```

## Initial Setup

### 1. Run the Setup Script

```bash
./setup.sh
```

This script will:
- Create necessary configuration directories
- Copy sample configuration files
- Set up shell integration
- Configure git integration

### 2. Manual Configuration

If you prefer manual setup:

1. **Create config directory:**
   ```bash
   mkdir -p ~/.config/gosh
   ```

2. **Copy sample configurations:**
   ```bash
   cp docs/sample.goshrc ~/.goshrc
   cp docs/sample.gosh_profile ~/.gosh_profile
   ```

3. **Edit configurations:**
   ```bash
   $EDITOR ~/.goshrc
   $EDITOR ~/.gosh_profile
   ```

## Configuration

### Configuration Files

Gosh uses two main configuration files:

- **`.gosh_profile`**: Loaded once at login (like `.bash_profile`)
- **`.goshrc`**: Loaded for each interactive shell (like `.bashrc`)

### Configuration Locations

Gosh looks for configuration files in this order:

1. `~/.config/gosh/goshrc`
2. `~/.goshrc`
3. `~/.gosh_profile` (for login shells)

### Environment Variables

Key environment variables for gosh:

```bash
# Prompt configuration
export GOSH_PROMPT_FORMAT="%u@%h:%w%g$ "
export GOSH_SHOW_GIT_INFO=true
export GOSH_PROMPT_COLOR=auto

# History configuration
export GOSH_HISTORY_SIZE=10000
export GOSH_HISTORY_FILE=~/.gosh_history
export GOSH_SAVE_HISTORY=true

# Completion configuration
export GOSH_COMPLETION_ENABLED=true
export GOSH_COMPLETION_CASE_INSENSITIVE=true

# Git integration
export GOSH_GIT_ENABLED=true
export GOSH_GIT_SHOW_STATUS=true
export GOSH_GIT_SHOW_BRANCH=true
```

## Shell Integration

### Set as Default Shell

⚠️ **Warning**: Only do this after thoroughly testing gosh!

1. **Add gosh to valid shells:**
   ```bash
   echo $(which gosh) | sudo tee -a /etc/shells
   ```

2. **Change default shell:**
   ```bash
   chsh -s $(which gosh)
   ```

### Terminal Integration

For terminal emulators that support it, you can configure gosh as the default shell:

- **iTerm2**: Preferences → Profiles → General → Command → Custom Shell
- **GNOME Terminal**: Preferences → Profiles → Command → Custom command
- **Alacritty**: Edit `~/.config/alacritty/alacritty.yml`

## Verification

### Test Basic Functionality

```bash
# Start gosh
gosh

# Test built-in commands
pwd
cd /tmp
help
history
alias
exit
```

### Test Git Integration

```bash
# In a git repository
gosh
# Your prompt should show git branch and status
```

### Test Tab Completion

```bash
# Press Tab after typing partial commands
ls <Tab>
git <Tab>
cd <Tab>
```

## Troubleshooting

### Common Issues

1. **Command not found:**
   - Ensure gosh is in your PATH
   - Check installation with `which gosh`

2. **Configuration not loading:**
   - Check file permissions: `ls -la ~/.goshrc`
   - Enable debug mode: `gosh --debug`

3. **Git integration not working:**
   - Ensure git is installed: `git --version`
   - Check git repository: `git status`

4. **Tab completion not working:**
   - Check completion settings in `.goshrc`
   - Verify GOSH_COMPLETION_ENABLED=true

### Debug Mode

Enable debug mode for troubleshooting:

```bash
gosh --debug
```

Or set in configuration:
```bash
export GOSH_DEBUG=true
```

### Log Files

Gosh logs can be found in:
- `~/.config/gosh/gosh.log` (if logging is enabled)
- System logs via `journalctl -u gosh` (if running as service)

## Uninstallation

### Remove Binary

```bash
sudo rm /usr/local/bin/gosh
```

### Remove Configuration

```bash
rm -rf ~/.config/gosh
rm ~/.goshrc ~/.gosh_profile ~/.gosh_history
```

### Restore Previous Shell

```bash
chsh -s /bin/bash  # or your previous shell
```

## Next Steps

- Read the [User Guide](USER_GUIDE.md) for detailed usage instructions
- Check out [Configuration Examples](CONFIGURATION.md) for advanced setups
- See [Development Guide](DEVELOPMENT.md) if you want to contribute

## Getting Help

- Check the [FAQ](FAQ.md)
- Open an issue on GitHub
- Join our community discussions
