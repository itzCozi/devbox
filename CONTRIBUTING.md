# Contributing to devbox

Thank you for your interest in contributing to devbox! This document provides guidelines and information for contributors.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [How to Contribute](#how-to-contribute)
- [Development Setup](#development-setup)
- [Coding Standards](#coding-standards)
- [Testing](#testing)
- [Submitting Changes](#submitting-changes)
- [Reporting Issues](#reporting-issues)
- [Feature Requests](#feature-requests)

## Code of Conduct

This project adheres to a code of conduct that we expect all contributors to follow. Please read [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) before contributing.

## Getting Started

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/devbox.git
   cd devbox
   ```
3. **Add the upstream repository**:
   ```bash
   git remote add upstream https://github.com/itzCozi/devbox.git
   ```

## How to Contribute

### Types of Contributions

We welcome several types of contributions:

- 🐛 **Bug fixes**
- ✨ **New features**
- 📚 **Documentation improvements**
- 🧪 **Tests**
- 🔧 **Code refactoring**
- 🎨 **UI/UX improvements**

### Good First Issues

Look for issues labeled with:
- `good first issue` - Perfect for newcomers
- `help wanted` - We'd love community help
- `documentation` - Improve our docs

## Development Setup

### Prerequisites

- Go 1.21 or later
- Docker
- Make
- Git

### Local Setup

1. **Install dependencies**:
   ```bash
   make deps
   ```

2. **Build the project**:
   ```bash
   make build
   ```

3. **Run tests**:
   ```bash
   make test
   ```

4. **Run all quality checks**:
   ```bash
   make ci
   ```

### Development Workflow

1. **Create a feature branch**:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes** following our coding standards

3. **Test your changes**:
   ```bash
   make ci
   ```

4. **Commit your changes**:
   ```bash
   git add .
   git commit -m "feat: add your feature description"
   ```

5. **Push to your fork**:
   ```bash
   git push origin feature/your-feature-name
   ```

6. **Create a Pull Request** on GitHub

## Coding Standards

### Go Code Style

- Follow standard Go formatting with `gofmt`
- Use `golangci-lint` for code quality
- Write meaningful variable and function names
- Add comments for exported functions and types
- Keep functions small and focused (max 50 statements)
- Maintain cyclomatic complexity under 15

### Code Organization

```
cmd/           # Main applications
internal/      # Private application code
docs/          # Documentation
scripts/       # Build and utility scripts
.github/       # GitHub workflows and templates
```

### Commit Message Format

We follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

```
<type>(<scope>): <description>

[optional body]

[optional footer(s)]
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

**Examples:**
```
feat(cli): add support for custom Docker registries
fix(config): resolve issue with JSON parsing
docs: update installation instructions
test(docker): add integration tests for container management
```

## Testing

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run tests with race detection
go test -race ./...
```

### Writing Tests

- Write unit tests for all new functions
- Include integration tests for CLI commands
- Aim for at least 50% test coverage
- Use table-driven tests when appropriate
- Mock external dependencies

### Test Structure

```go
func TestFunctionName(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        // test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test implementation
        })
    }
}
```

## Submitting Changes

### Pull Request Process

1. **Update documentation** if needed
2. **Add tests** for new functionality
3. **Ensure CI passes** - all checks must be green
4. **Update CHANGELOG.md** if applicable
5. **Request review** from maintainers

### Pull Request Guidelines

- **Title**: Use a descriptive title following conventional commits
- **Description**: 
  - Explain what changes you made and why
  - Reference any related issues
  - Include screenshots for UI changes
  - List any breaking changes

### Pull Request Template

Your PR description should include:

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Tests pass locally
- [ ] Added tests for new functionality
- [ ] Updated documentation

## Checklist
- [ ] Code follows style guidelines
- [ ] Self-review completed
- [ ] Comments added for complex code
- [ ] Documentation updated
- [ ] No breaking changes (or clearly documented)
```

## Reporting Issues

### Bug Reports

When reporting bugs, please include:

1. **Clear title** describing the issue
2. **Environment details**:
   - OS and version
   - Go version
   - Docker version
   - devbox version
3. **Steps to reproduce**
4. **Expected behavior**
5. **Actual behavior**
6. **Error messages or logs**
7. **Additional context**

### Security Issues

For security vulnerabilities, please email [security@yourdomain.com] instead of creating a public issue.

## Feature Requests

We welcome feature requests! Please:

1. **Check existing issues** to avoid duplicates
2. **Describe the problem** you're trying to solve
3. **Propose a solution** if you have one
4. **Consider the scope** - start small and iterate
5. **Be open to discussion** about implementation

## Documentation

### Documentation Types

- **API documentation**: Inline code comments
- **User guides**: Markdown files in `docs/`
- **README updates**: For setup and usage
- **Changelog**: Track notable changes

### Documentation Standards

- Write clear, concise prose
- Include code examples
- Test all examples
- Keep documentation up-to-date with code changes

## Community

### Communication Channels

- **GitHub Issues**: Bug reports, feature requests
- **GitHub Discussions**: General questions, ideas
- **Pull Requests**: Code review and discussion

### Getting Help

- Check existing documentation
- Search GitHub issues
- Create a new issue with the `question` label
- Join community discussions

## Recognition

Contributors will be recognized in:
- GitHub contributors list
- CHANGELOG.md for significant contributions
- Special thanks in release notes

## License

By contributing to devbox, you agree that your contributions will be licensed under the MIT License.

---

Thank you for contributing to devbox! 🚀