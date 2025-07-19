# Security Policy

## Supported Versions

We support the latest released version of go-libconfig with security updates.

| Version | Supported          |
| ------- | ------------------ |
| Latest  | :white_check_mark: |
| < Latest| :x:                |

## Reporting a Vulnerability

If you discover a security vulnerability, please report it to us responsibly.

### How to Report

**Please do NOT report security vulnerabilities through public GitHub issues.**

Instead, please report security vulnerabilities by:

1. **GitHub Security Advisories**: Use the ["Report a vulnerability"](https://github.com/kuzmik/go-libconfig/security/advisories/new) feature on GitHub
2. **Email**: Send details to [security@your-domain.com] (replace with your actual email)

### What to Include

When reporting a vulnerability, please include:

- **Description**: A clear description of the vulnerability
- **Impact**: What could an attacker achieve?
- **Reproduction**: Steps to reproduce the vulnerability
- **Affected versions**: Which versions are affected
- **Suggested fix**: If you have ideas for how to fix it
- **Your contact info**: How we can reach you for follow-up

### Response Timeline

- **Acknowledgment**: We will acknowledge receipt within 48 hours
- **Initial assessment**: We will provide an initial assessment within 5 business days
- **Resolution**: We aim to resolve critical vulnerabilities within 30 days
- **Disclosure**: We will coordinate disclosure timing with you

### Security Best Practices

When using go-libconfig:

1. **Input Validation**: Always validate configuration files from untrusted sources
2. **File Permissions**: Restrict access to configuration files containing sensitive data
3. **Include Paths**: Be careful with `@include` directives and untrusted file paths
4. **Error Handling**: Use static error types for consistent error handling
5. **Dependencies**: Keep go-libconfig updated to the latest version

### Known Security Considerations

- **File Inclusion**: The `@include` directive can read arbitrary files if not properly controlled
- **Path Traversal**: Include paths should be validated to prevent directory traversal attacks
- **Memory Usage**: Very large configuration files could cause memory exhaustion
- **Parsing Complexity**: Deeply nested structures could cause stack overflow

### Security Features

- **Static Error Types**: Prevents error injection attacks
- **Input Validation**: Comprehensive validation throughout parsing
- **No External Commands**: The library doesn't execute external commands
- **Memory Safe**: Pure Go implementation with bounds checking
- **Include Depth Limiting**: Prevents infinite recursion in includes

## Vulnerability Disclosure Policy

We believe in coordinated disclosure:

1. **Private disclosure**: Report the issue privately first
2. **Investigation**: We investigate and develop a fix
3. **Coordination**: We coordinate with you on disclosure timing
4. **Public disclosure**: After a fix is available, we publicly disclose the issue
5. **Credit**: We provide credit to security researchers (unless they prefer to remain anonymous)

## Security Updates

Security updates will be released as patch versions and announced through:

- GitHub Security Advisories
- GitHub Releases
- Project README and CHANGELOG

Thank you for helping keep go-libconfig secure!