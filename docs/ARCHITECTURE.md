# Gosh Architecture

This document describes the architecture and design principles of the Gosh shell.

## Overview

Gosh is designed as a modular, extensible shell with clear separation of concerns. The architecture follows Go best practices and emphasizes maintainability, testability, and performance.

## Core Principles

1. **Modularity**: Each component has a single responsibility
2. **Testability**: All components are designed to be easily testable
3. **Performance**: Efficient execution with minimal overhead
4. **Extensibility**: Easy to add new features and commands
5. **Compatibility**: Bash-like behavior where appropriate

## Project Structure

```
gosh/
├── cmd/                    # Application entry points
│   └── main.go            # Main application entry point
├── internal/              # Internal packages (not importable)
│   ├── shell/             # Core shell functionality
│   ├── parser/            # Command parsing and execution
│   ├── completion/        # Tab completion system
│   ├── prompt/            # Prompt generation and customization
│   ├── config/            # Configuration management
│   ├── git/               # Git integration
│   └── history/           # Command history management
├── docs/                  # Documentation
├── scripts/               # Utility scripts
└── Makefile              # Build automation
```

## Component Architecture

### 1. Shell Core (`internal/shell`)

The shell core is the main orchestrator that ties all components together.

**Key Components:**
- `Shell`: Main shell instance that manages the REPL loop
- Signal handling for graceful shutdown
- Configuration loading and management
- Component initialization and coordination

**Responsibilities:**
- Initialize all subsystems
- Manage the main read-eval-print loop
- Handle signals (SIGINT, SIGTERM)
- Coordinate between different components

### 2. Parser (`internal/parser`)

The parser handles command line parsing and execution.

**Key Components:**
- `Parser`: Main parsing engine
- `Command`: Interface for all commands
- Built-in command implementations
- External command execution

**Responsibilities:**
- Tokenize input with proper quote handling
- Expand aliases
- Identify built-in vs external commands
- Execute commands with proper context

**Command Types:**
- **Built-in Commands**: `cd`, `pwd`, `exit`, `help`, `history`, `alias`, `export`
- **External Commands**: System commands executed via `exec`
- **No-op Commands**: Empty input handling

### 3. Completion System (`internal/completion`)

Provides intelligent tab completion functionality.

**Key Components:**
- `Manager`: Main completion coordinator
- Command completion logic
- File/directory completion
- Context-aware suggestions

**Completion Types:**
- **Command Completion**: Built-ins, aliases, PATH commands
- **File Completion**: Files and directories with filtering
- **Git Completion**: Branches, remotes, modified files
- **Context-Aware**: Different completions based on command context

### 4. Prompt System (`internal/prompt`)

Generates customizable prompts with git integration.

**Key Components:**
- `Manager`: Prompt generation coordinator
- Format string parsing
- Git information integration
- Color scheme application

**Prompt Features:**
- **Format Codes**: `%u` (user), `%h` (host), `%w` (working dir), `%g` (git info)
- **Git Integration**: Branch, status, ahead/behind indicators
- **Color Schemes**: Auto, minimal, bright, none
- **Dynamic Updates**: Real-time git status updates

### 5. Configuration System (`internal/config`)

Manages configuration loading and parsing.

**Key Components:**
- `Config`: Configuration structure
- File parsing (`.goshrc`, `.gosh_profile`)
- Environment variable handling
- Default configuration generation

**Configuration Sources:**
1. `~/.config/gosh/goshrc`
2. `~/.goshrc`
3. `~/.gosh_profile` (login shells)
4. Environment variables
5. Command-line flags

### 6. Git Integration (`internal/git`)

Provides git repository information and integration.

**Key Components:**
- `Manager`: Git operations coordinator
- `Info`: Git repository information structure
- Repository detection and status checking
- Branch and remote information

**Git Features:**
- **Repository Detection**: Automatic git repo detection
- **Status Information**: Uncommitted, untracked, staged files
- **Branch Information**: Current branch or commit hash
- **Ahead/Behind**: Tracking branch comparison
- **Completion Support**: Branches, remotes, modified files

### 7. History Management (`internal/history`)

Manages command history with persistence and search.

**Key Components:**
- `Manager`: History operations coordinator
- `Entry`: Individual history entry structure
- Persistent storage
- Search and navigation

**History Features:**
- **Persistent Storage**: Save/load from file
- **Search Capabilities**: Text search and prefix matching
- **Navigation**: Previous/next command navigation
- **Deduplication**: Optional duplicate removal
- **Size Management**: Configurable history size limits

## Data Flow

### 1. Shell Startup

```
main() → Config Loading → Shell Creation → Component Initialization → Main Loop
```

1. **Configuration Loading**: Load from various sources
2. **Shell Creation**: Initialize shell with configuration
3. **Component Initialization**: Create managers for each subsystem
4. **Main Loop**: Start the REPL loop

### 2. Command Execution

```
Input → Tokenization → Alias Expansion → Command Identification → Execution
```

1. **Input Reading**: Read line from user
2. **Tokenization**: Parse into tokens with quote handling
3. **Alias Expansion**: Expand any aliases
4. **Command Identification**: Built-in vs external
5. **Execution**: Execute with proper context

### 3. Prompt Generation

```
Prompt Request → Format Parsing → Information Gathering → Color Application → Display
```

1. **Format Parsing**: Parse format string
2. **Information Gathering**: Collect user, host, directory, git info
3. **Color Application**: Apply color scheme
4. **Display**: Output formatted prompt

### 4. Tab Completion

```
Tab Press → Context Analysis → Completion Generation → Display/Application
```

1. **Context Analysis**: Determine what to complete
2. **Completion Generation**: Generate appropriate completions
3. **Display/Application**: Show options or apply completion

## Design Patterns

### 1. Manager Pattern

Each major subsystem uses a Manager struct that:
- Encapsulates related functionality
- Maintains necessary state
- Provides a clean interface
- Handles initialization and cleanup

### 2. Command Pattern

Commands implement a common interface:
```go
type Command interface {
    Execute(ctx context.Context, cfg *config.Config) error
}
```

This allows:
- Uniform command execution
- Easy addition of new commands
- Proper context and configuration passing
- Testable command implementations

### 3. Strategy Pattern

Different completion strategies based on context:
- Command completion for first token
- File completion for arguments
- Git-specific completion for git commands

### 4. Observer Pattern

Components can react to configuration changes:
- Prompt updates when git settings change
- Completion behavior updates when settings change
- History behavior updates when settings change

## Error Handling

### Principles

1. **Graceful Degradation**: Continue operation when possible
2. **User-Friendly Messages**: Clear, actionable error messages
3. **Debug Information**: Detailed errors in debug mode
4. **Recovery**: Attempt to recover from non-fatal errors

### Error Categories

- **Configuration Errors**: Invalid config files, missing settings
- **Command Errors**: Command not found, execution failures
- **System Errors**: File system, permission issues
- **Git Errors**: Repository issues, git command failures

## Performance Considerations

### Optimization Strategies

1. **Lazy Loading**: Load components only when needed
2. **Caching**: Cache expensive operations (git status, completions)
3. **Efficient Parsing**: Minimal allocations in hot paths
4. **Concurrent Operations**: Use goroutines for independent operations

### Memory Management

- **Bounded History**: Limit history size to prevent memory growth
- **Completion Caching**: Cache completion results with TTL
- **String Interning**: Reuse common strings where beneficial

## Testing Strategy

### Unit Tests

- Each component has comprehensive unit tests
- Mock external dependencies (git, filesystem)
- Test error conditions and edge cases

### Integration Tests

- Test component interactions
- End-to-end command execution
- Configuration loading and application

### Benchmarks

- Performance tests for critical paths
- Memory allocation tracking
- Comparison with baseline implementations

## Future Architecture Considerations

### Plugin System

Potential architecture for plugins:
- Plugin interface definition
- Dynamic loading mechanism
- Sandboxing and security
- Plugin discovery and management

### Scripting Support

Enhanced scripting capabilities:
- Script parsing and execution
- Variable scoping
- Control flow structures
- Function definitions

### Remote Shell Support

Architecture for remote shell capabilities:
- Client-server communication
- Session management
- Security and authentication
- State synchronization

## Contributing to Architecture

When making architectural changes:

1. **Document Changes**: Update this document
2. **Maintain Interfaces**: Preserve existing interfaces when possible
3. **Add Tests**: Include tests for new components
4. **Consider Performance**: Profile changes for performance impact
5. **Review Dependencies**: Minimize external dependencies

## Conclusion

The Gosh architecture is designed to be modular, testable, and extensible. Each component has clear responsibilities and well-defined interfaces. This design enables easy maintenance, testing, and future enhancements while providing a solid foundation for a modern shell implementation.
