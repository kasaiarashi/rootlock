# Contributing to RootLock

Thank you for considering contributing to RootLock! This document provides guidelines and instructions for contributing.

## Code of Conduct

This project adheres to a Code of Conduct. By participating, you are expected to uphold this code.

## How to Contribute

### Reporting Bugs

1. **Check existing issues** to avoid duplicates
2. **Create a new issue** with:
   - Clear, descriptive title
   - Steps to reproduce
   - Expected vs actual behavior
   - Environment details (OS, version, etc.)
   - Screenshots if applicable

### Suggesting Features

1. **Search existing feature requests**
2. **Create a new issue** with:
   - Use case / problem being solved
   - Proposed solution
   - Alternative approaches considered

### Security Vulnerabilities

**DO NOT** open public issues for security vulnerabilities.

Email: **security@rootlock.kriaa.in**

## Development Workflow

### 1. Fork & Clone

```bash
git clone https://github.com/YOUR_USERNAME/rootlock.git
cd rootlock
git remote add upstream https://github.com/rootlock/rootlock.git
```

### 2. Create Branch

```bash
git checkout dev
git pull upstream dev
git checkout -b feature/your-feature-name
```

**Branch naming**:
- `feature/` - New features
- `fix/` - Bug fixes
- `docs/` - Documentation updates
- `refactor/` - Code refactoring
- `test/` - Test improvements

### 3. Make Changes

- Write clean, idiomatic code
- Add tests for new functionality
- Update documentation as needed
- Follow existing code style

### 4. Test

```bash
# Backend
cd backend && go test ./...

# Desktop
cd desktop && npm test

# iOS
open mobile/ios/RootLock.xcodeproj
# Run tests in Xcode (⌘U)
```

### 5. Commit

Use descriptive commit messages:

```
Add passkey authentication to iOS app

- Implement ASAuthorizationController
- Add biometric unlock flow
- Update UI for passkey management

Fixes #123
```

**Format**:
- First line: Summary (50 chars max)
- Blank line
- Detailed description (72 chars per line)
- Reference issues/PRs

### 6. Push & Pull Request

```bash
git push origin feature/your-feature-name
```

Open a Pull Request against `dev` branch.

**PR Template**:
- Description of changes
- Related issues
- Testing performed
- Screenshots (if UI changes)
- Checklist completed

## Development Setup

See [docs/development.md](docs/development.md) for detailed setup instructions.

## Code Style

### Go
- Use `gofmt` and `goimports`
- Follow [Effective Go](https://go.dev/doc/effective_go)
- Run `golangci-lint`

### Swift
- Use SwiftLint
- Follow [Swift API Design Guidelines](https://www.swift.org/documentation/api-design-guidelines/)

### Rust
- Use `rustfmt` and `clippy`
- Follow [Rust API Guidelines](https://rust-lang.github.io/api-guidelines/)

### TypeScript/JavaScript
- Use ESLint + Prettier
- Follow Airbnb style guide

## Testing Requirements

- All new features must include tests
- Maintain >80% code coverage for critical paths
- Security-related code requires comprehensive tests

## Documentation

Update documentation for:
- New features
- API changes
- Configuration options
- Breaking changes

## Review Process

1. Automated checks must pass (CI/CD)
2. Code review by maintainer(s)
3. Requested changes addressed
4. Approval → Merge to `dev`

## Release Process

1. Features merged to `dev`
2. Testing period on `dev`
3. Create `release` branch
4. Final testing
5. Merge to `master`
6. Tag version (`v1.0.0`)
7. Publish release

## Questions?

- **General**: Open a GitHub Discussion
- **Bugs**: GitHub Issues
- **Security**: security@rootlock.kriaa.in

---

Thank you for contributing! 🎉
