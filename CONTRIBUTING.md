# Contributing to Gosh

Thank you for your interest in contributing to Gosh! This document provides guidelines and information for contributors.

## Getting Started

### Prerequisites

- Go 1.19 or later
- Git
- Make (for build automation)
- A Unix-like operating system (Linux, macOS, BSD)

### Setting Up Development Environment

1. **Fork and clone the repository:**
   ```bash
   git clone https://github.com/yourusername/gosh.git
   cd gosh
   ```

2. **Install dependencies:**
   ```bash
   make deps
   ```

3. **Build the project:**
   ```bash
   make build
   ```

4. **Run tests:**
   ```bash
   make test
   ```

## Development Workflow

### Making Changes

1. **Create a feature branch:**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes:**
   - Follow the coding standards (see below)
   - Add tests for new functionality
   - Update documentation as needed

3. **Test your changes:**
   ```bash
   make test
   make lint
   make vet
   ```

4. **Commit your changes:**
   ```bash
   git add .
   git commit -m "feat: add your feature description"
   ```

5. **Push and create a pull request:**
   ```bash
   git push origin feature/your-feature-name
   ```

### Commit Message Format

We follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

Types:
- `feat`: A new feature
- `fix`: A bug fix
- `docs`: Documentation only changes
- `style`: Changes that do not affect the meaning of the code
- `refactor`: A code change that neither fixes a bug nor adds a feature
- `perf`: A code change that improves performance
- `test`: Adding missing tests or correcting existing tests
- `chore`: Changes to the build process or auxiliary tools

Examples:
```
feat(completion): add git branch completion
fix(parser): handle quoted arguments correctly
docs: update installation instructions
```

## Coding Standards

### Go Code Style

- Follow the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` to format your code
- Run `go vet` to check for common mistakes
- Use meaningful variable and function names
- Add comments for exported functions and types

### Code Organization

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
└── scripts/       # Utility scripts
```

### Documentation

- Document all exported functions and types
- Include examples in documentation
- Update README.md for significant changes
- Add or update relevant documentation in `docs/`

### Testing

- Write unit tests for new functionality
- Aim for good test coverage
- Use table-driven tests where appropriate
- Include integration tests for complex features

Example test structure:
```go
func TestFunctionName(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        {
            name:     "valid input",
            input:    "test",
            expected: "expected output",
            wantErr:  false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := FunctionName(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("FunctionName() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if result != tt.expected {
                t.Errorf("FunctionName() = %v, want %v", result, tt.expected)
            }
        })
    }
}
```

## Areas for Contribution

### High Priority

- **Tab Completion Enhancements**: Improve completion for various commands
- **History Search**: Implement Ctrl+R style history search
- **Configuration Loading**: Improve configuration file parsing
- **Error Handling**: Better error messages and recovery
- **Performance**: Optimize shell startup and command execution

### Medium Priority

- **Syntax Highlighting**: Add syntax highlighting for commands
- **Auto-suggestions**: Fish-like auto-suggestions
- **Plugin System**: Architecture for extending gosh
- **Themes**: Customizable color themes
- **Scripting**: Better support for shell scripts

### Low Priority

- **Windows Support**: Port to Windows
- **Job Control**: Background job management
- **Advanced Redirections**: More redirection operators
- **Globbing**: Pattern matching for file names

## Pull Request Process

1. **Ensure your PR:**
   - Has a clear description of the changes
   - Includes tests for new functionality
   - Updates documentation as needed
   - Passes all existing tests
   - Follows the coding standards

2. **PR Review Process:**
   - At least one maintainer must approve the PR
   - All CI checks must pass
   - Address any feedback from reviewers

3. **After Approval:**
   - Maintainers will merge the PR
   - Your contribution will be included in the next release

## Reporting Issues

### Bug Reports

When reporting bugs, please include:

- **Description**: Clear description of the issue
- **Steps to Reproduce**: Detailed steps to reproduce the bug
- **Expected Behavior**: What you expected to happen
- **Actual Behavior**: What actually happened
- **Environment**: OS, Go version, gosh version
- **Additional Context**: Any other relevant information

### Feature Requests

When requesting features, please include:

- **Description**: Clear description of the feature
- **Use Case**: Why this feature would be useful
- **Proposed Implementation**: Ideas for how it could be implemented
- **Alternatives**: Any alternative solutions you've considered

## Code of Conduct

### Our Pledge

We are committed to making participation in this project a harassment-free experience for everyone, regardless of age, body size, disability, ethnicity, gender identity and expression, level of experience, nationality, personal appearance, race, religion, or sexual identity and orientation.

### Our Standards

Examples of behavior that contributes to creating a positive environment include:

- Using welcoming and inclusive language
- Being respectful of differing viewpoints and experiences
- Gracefully accepting constructive criticism
- Focusing on what is best for the community
- Showing empathy towards other community members

### Enforcement

Instances of abusive, harassing, or otherwise unacceptable behavior may be reported by contacting the project team. All complaints will be reviewed and investigated and will result in a response that is deemed necessary and appropriate to the circumstances.

## Getting Help

- **Documentation**: Check the `docs/` directory
- **Issues**: Search existing issues on GitHub
- **Discussions**: Use GitHub Discussions for questions
- **Community**: Join our community channels (if available)

## Recognition

Contributors will be recognized in:

- The project's README.md
- Release notes for significant contributions
- The project's contributors page

Thank you for contributing to Gosh!
