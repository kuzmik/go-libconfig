# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial release of go-libconfig library
- Full libconfig specification support for scalars, arrays, groups, and lists
- Multiple integer formats: decimal, hexadecimal, binary, and octal
- String features including escape sequences, concatenation, and Unicode support
- Include directives (`@include`) for modular configurations
- Comprehensive error handling with static error types
- Type-safe value lookup methods
- Extensive test suite with >95% coverage
- Benchmark tests for performance monitoring
- GitHub Actions CI/CD pipeline
- Golangci-lint configuration for code quality
- MIT license for commercial use

### Security
- Static error types prevent error injection attacks
- Input validation and sanitization throughout parsing
- No unsafe operations or external command execution

## Release Guidelines

### Version Numbering
- **MAJOR**: Incompatible API changes
- **MINOR**: New functionality in a backwards compatible manner
- **PATCH**: Backwards compatible bug fixes

### Release Process
1. Update this CHANGELOG.md with the new version
2. Create a new git tag: `git tag v1.0.0`
3. Push the tag: `git push origin v1.0.0`
4. GitHub Actions will automatically create the release