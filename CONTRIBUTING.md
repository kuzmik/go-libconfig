# Contributing to go-libconfig

Thank you for your interest in contributing to go-libconfig! This document provides guidelines and information for contributors.

## Code of Conduct

This project follows the [Go Community Code of Conduct](https://go.dev/conduct). Please be respectful and inclusive in all interactions.

## Getting Started

### Prerequisites

- Go 1.21 or higher
- Git
- Make (for running development tasks)

### Development Setup

1. **Fork and clone the repository:**
   ```bash
   git clone https://github.com/YOUR_USERNAME/go-libconfig.git
   cd go-libconfig
   ```

2. **Install dependencies:**
   ```bash
   go mod download
   ```

3. **Run tests to verify setup:**
   ```bash
   make test
   ```

4. **Run all checks:**
   ```bash
   make check
   ```

## Development Workflow

### Making Changes

1. **Create a feature branch:**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes:**
   - Write code following Go conventions
   - Add tests for new functionality
   - Update documentation as needed
   - Follow the existing code style

3. **Test your changes:**
   ```bash
   # Run tests
   make test

   # Run tests with race detection
   make race

   # Run linting
   make lint

   # Run benchmarks
   make bench

   # Generate coverage report
   make coverage
   ```

4. **Commit your changes:**
   ```bash
   git add .
   git commit -m "feat: add new libconfig feature"
   ```

### Commit Message Format

We follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

**Types:**
- `feat`: New features
- `fix`: Bug fixes
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

**Examples:**
```
feat: add support for scientific notation in floats
fix: handle edge case in string parsing
docs: update API reference for new methods
test: add benchmark for large array parsing
```

## Code Guidelines

### Go Style

- Follow [Effective Go](https://go.dev/doc/effective_go)
- Use `gofmt` for formatting
- Follow the [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments)
- Use meaningful variable and function names
- Add godoc comments for exported functions

### Testing

- **Test Coverage**: Maintain >95% test coverage
- **Test Names**: Use descriptive test names (`TestParseComplexConfiguration`)
- **Table Tests**: Use table-driven tests for multiple test cases
- **Edge Cases**: Include tests for edge cases and error conditions
- **Benchmarks**: Add benchmarks for performance-critical code

### Error Handling

- Use static error types (e.g., `ErrSettingNotFound`)
- Wrap errors with context using `fmt.Errorf("context: %w", err)`
- Return meaningful error messages with line/column information where applicable

### Documentation

- Add godoc comments for all exported types, functions, and methods
- Update README.md for new features
- Include code examples in documentation
- Update CHANGELOG.md for notable changes

## Pull Requests

### Before Submitting

1. **Ensure all checks pass:**
   ```bash
   make check
   ```

2. **Update documentation:**
   - Update README.md if needed
   - Add godoc comments
   - Update examples if applicable

3. **Add tests:**
   - Unit tests for new functionality
   - Integration tests for complex features
   - Benchmark tests for performance-critical code

### PR Guidelines

- **Title**: Use descriptive titles that explain the change
- **Description**: Include:
  - What changes were made
  - Why the changes were necessary
  - Any breaking changes
  - Testing performed
- **Size**: Keep PRs focused and reasonably sized
- **Reviews**: Be responsive to feedback and suggestions

### PR Template

Your PR should include:

```markdown
## Summary
Brief description of changes

## Changes Made
- List of specific changes
- Any new features added
- Any bugs fixed

## Testing
- [ ] All existing tests pass
- [ ] New tests added for new functionality
- [ ] Benchmarks updated if applicable
- [ ] Manual testing performed

## Documentation
- [ ] README updated if needed
- [ ] Godoc comments added/updated
- [ ] Examples updated if needed

## Breaking Changes
List any breaking changes and migration path

## Additional Notes
Any additional context or notes for reviewers
```

## Reporting Issues

### Bug Reports

When reporting bugs, include:

1. **Go version**: `go version`
2. **OS/Platform**: Linux, macOS, Windows
3. **Minimal reproduction case**
4. **Expected vs actual behavior**
5. **Error messages** (if any)

### Feature Requests

For feature requests, include:

1. **Use case**: Why is this feature needed?
2. **Proposed solution**: How should it work?
3. **Alternatives considered**: Other approaches you've considered
4. **Examples**: Code examples of how it would be used

## Release Process

Releases are automated through GitHub Actions:

1. **Update CHANGELOG.md** with new version
2. **Create and push a version tag:**
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```
3. **GitHub Actions** automatically creates the release

## Getting Help

- **Documentation**: Check the [README](README.md) and godoc
- **Issues**: Search existing issues before creating new ones
- **Discussions**: Use GitHub Discussions for questions and ideas

## Recognition

Contributors will be recognized in release notes and the project README. Thank you for helping make go-libconfig better!