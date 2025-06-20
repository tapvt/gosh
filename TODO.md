# Gosh Shell - TODO Punch List

This document outlines the tasks needed to complete the gosh shell implementation. Items are prioritized by importance and complexity.

## 🔴 Critical Issues (Must Fix)

### 1. Configuration File Execution
- **Issue**: `loadConfigFile()` in `internal/shell/shell.go:267-271` is a stub implementation
- **Status**: Not implemented - just returns `nil`
- **Impact**: Configuration files (.goshrc, .gosh_profile) are not actually executed
- **Action**: Implement proper config file execution that runs shell commands from config files

### 2. Test Failures
- **Issue**: `TestLoad_NonExistentFile` in `internal/config/config_test.go:328` is failing
- **Status**: Test expects `os.IsNotExist` error but gets `nil` due to fallback to home directory .goshrc
- **Action**: Fix test logic or Load function behavior to match expected semantics

### 3. Missing Test Coverage
- **Issue**: No test files for `internal/git/git.go` and `internal/prompt/prompt.go`
- **Status**: Missing test coverage for critical components
- **Action**: Create `git_test.go` and `prompt_test.go` with comprehensive test suites

## 🟡 High Priority Features

### 4. History Search (Ctrl+R)
- **Issue**: No reverse history search implementation
- **Status**: Documented in CONTRIBUTING.md as high priority but not implemented
- **Action**: Implement Ctrl+R style history search functionality

### 5. Advanced Shell Features
- **Issue**: Missing core shell features mentioned in documentation
- **Status**: Documented but not implemented
- **Missing Features**:
  - Pipes (`|`) - parser mentions pipes but no implementation
  - Redirections (`>`, `>>`, `<`) - parser mentions redirections but no implementation
  - Background processes (`&`)
  - Command substitution (`$(command)` or `` `command` ``)
  - Job control (jobs, fg, bg)

### 6. Enhanced Built-in Commands
- **Issue**: Limited built-in command set
- **Status**: Basic commands implemented, missing common ones
- **Missing Commands**:
  - `source` / `.` - Execute commands from file
  - `which` - Locate command
  - `type` - Display command type
  - `jobs`, `fg`, `bg` - Job control
  - `set`, `unset` - Shell options and variables
  - `echo` - Built-in echo (currently relies on external)

## 🟢 Medium Priority Enhancements

### 7. Git Integration Improvements
- **Issue**: Git integration exists but could be enhanced
- **Status**: Basic git info in prompt, needs more features
- **Enhancements**:
  - Git command completion for branch names, remotes, etc.
  - Better git status indicators
  - Git hooks integration
  - Performance optimization for large repositories

### 8. Tab Completion Enhancements
- **Issue**: Basic completion works but needs improvement
- **Status**: File and command completion implemented
- **Enhancements**:
  - Context-aware completion for specific commands
  - Completion for environment variables
  - Completion for aliases
  - Smart completion for paths with spaces
  - Completion caching for performance

### 9. Prompt System Improvements
- **Issue**: Basic prompt formatting implemented
- **Status**: Works but limited customization
- **Enhancements**:
  - More format codes (%d for date, %T for time, etc.)
  - Color customization per element
  - Conditional prompt elements
  - Multi-line prompt support
  - Right-side prompt (RPROMPT)

### 10. Configuration System Enhancements
- **Issue**: Configuration loading works but limited functionality
- **Status**: Parses config files but doesn't execute shell commands
- **Enhancements**:
  - Support for functions in config files
  - Conditional configuration loading
  - Configuration validation
  - Runtime configuration changes
  - Configuration file templates

## 🔵 Low Priority / Nice to Have

### 11. Advanced Features
- **Issue**: Missing modern shell conveniences
- **Status**: Not implemented
- **Features**:
  - Syntax highlighting
  - Auto-suggestions (fish-like)
  - Fuzzy completion
  - Plugin system
  - Themes and color schemes
  - Command timing and performance metrics

### 12. Error Handling Improvements
- **Issue**: Basic error handling exists but could be better
- **Status**: Some error categorization implemented
- **Improvements**:
  - Better error messages with suggestions
  - Error recovery mechanisms
  - Logging system
  - Debug mode enhancements

### 13. Performance Optimizations
- **Issue**: No performance optimizations implemented
- **Status**: Basic functionality works
- **Optimizations**:
  - Startup time optimization
  - Memory usage optimization
  - Completion caching
  - Git status caching
  - Lazy loading of components

### 14. Documentation and Examples
- **Issue**: Good documentation exists but could be expanded
- **Status**: Basic docs in place
- **Additions**:
  - More configuration examples
  - Scripting guide
  - Migration guide from bash/zsh
  - Performance tuning guide
  - Troubleshooting guide

## 🛠️ Development Infrastructure

### 15. CI/CD Pipeline
- **Issue**: No automated testing/deployment
- **Status**: Manual testing only
- **Action**: Set up GitHub Actions for automated testing, linting, and releases

### 16. Release Management
- **Issue**: No formal release process
- **Status**: Manual builds only
- **Action**: Implement semantic versioning, automated releases, and distribution

### 17. Benchmarking
- **Issue**: No performance benchmarks
- **Status**: Makefile has bench target but no benchmarks implemented
- **Action**: Add performance benchmarks for critical paths

## 📋 Quick Wins (Easy Fixes)

### 18. Fix README GitHub URL
- **Issue**: README.md line 24 and 135 have placeholder GitHub URL
- **Status**: Shows "yourusername" instead of actual repository
- **Action**: Update to use actual repository URL (git@github.com:tapvt/gosh.git)

### 19. Add Missing Sample Files
- **Issue**: Setup script references sample files that may not exist
- **Status**: `docs/sample.gosh_profile` exists, verify others
- **Action**: Ensure all referenced sample files exist and are complete

### 20. Improve Makefile
- **Issue**: Some Makefile targets could be enhanced
- **Status**: Comprehensive Makefile exists
- **Improvements**:
  - Add target for running specific tests
  - Add target for generating test coverage reports
  - Add target for cross-compilation testing

---

## Priority Order for Implementation

1. **Fix configuration file execution** (Critical for basic functionality)
2. **Fix failing tests** (Critical for development workflow)
3. **Add missing test coverage** (Critical for code quality)
4. **Implement pipes and redirections** (High priority shell features)
5. **Add history search** (High priority user experience)
6. **Enhance built-in commands** (Medium priority functionality)
7. **Improve git integration** (Medium priority features)
8. **Add advanced shell features** (Low priority enhancements)

This TODO list provides a roadmap for completing the gosh shell implementation, focusing on critical functionality first, then user experience improvements, and finally advanced features.
