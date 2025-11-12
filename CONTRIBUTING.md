# Contributing to Vault Watcher

Thank you for your interest in contributing to Vault Watcher! This document provides guidelines and information for contributors.

## Getting Started

### Prerequisites

- Go 1.21 or later
- Git
- HashiCorp Vault (for integration tests)

### Setting Up Development Environment

1. Fork the repository on GitHub
2. Clone your fork locally:
   ```bash
   git clone https://github.com/your-username/vault-watcher.git
   cd vault-watcher
   ```

3. Install dependencies:
   ```bash
   go mod download
   ```

4. Verify everything works:
   ```bash
   go test -v
   ```

## Development Workflow

### Making Changes

1. Create a new branch for your feature/fix:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. Make your changes following the coding standards below

3. Add or update tests as needed

4. Run the test suite:
   ```bash
   go test -v -race ./...
   ```

5. Run linting:
   ```bash
   golangci-lint run
   ```

6. Commit your changes with a descriptive message:
   ```bash
   git commit -m "feat: add new feature description"
   ```

### Coding Standards

- Follow standard Go formatting (`go fmt`)
- Write comprehensive tests for new functionality
- Add godoc comments for exported functions and types
- Keep functions focused and small
- Use meaningful variable and function names
- Handle errors appropriately

### Commit Message Format

We follow conventional commits format:

- `feat`: New features
- `fix`: Bug fixes
- `docs`: Documentation changes
- `test`: Adding or updating tests
- `refactor`: Code refactoring
- `chore`: Maintenance tasks

Examples:
- `feat: add support for KV v2 engine`
- `fix: handle nil vault response correctly`
- `docs: update installation instructions`

## Testing

### Unit Tests

Run unit tests with:
```bash
go test -v
```

With coverage:
```bash
go test -v -cover
go test -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Integration Tests

Integration tests require a running Vault instance:

```bash
# Start Vault dev server
vault server -dev

# Set environment variables
export VAULT_ADDR='http://127.0.0.1:8200'
export VAULT_HOST='http://127.0.0.1:8200'
export VAULT_PATH='secret/test'
export VAULT_TOKEN='your-dev-token'

# Create test data
vault kv put secret/test key1=value1 key2=value2

# Run integration tests
go test -tags=integration -v
```

### Test Guidelines

- Write tests for all new functionality
- Test both success and error cases
- Use table-driven tests where appropriate
- Mock external dependencies when possible
- Ensure tests are deterministic and can run in parallel

## Submitting Changes

### Pull Request Process

1. Push your branch to your fork:
   ```bash
   git push origin feature/your-feature-name
   ```

2. Create a pull request on GitHub with:
   - Clear title and description
   - Reference to any related issues
   - Screenshots/examples if applicable

3. Ensure all CI checks pass:
   - Tests pass
   - Linting passes
   - No security issues

4. Respond to review feedback promptly

### Pull Request Requirements

- [ ] Tests added/updated and passing
- [ ] Documentation updated if needed
- [ ] Changelog entry added (if applicable)
- [ ] Code follows project conventions
- [ ] Commit messages follow conventional format

## Code Review Process

- All submissions require review before merging
- We aim to review PRs within 48 hours
- Address reviewer feedback promptly
- Maintain a professional and respectful tone

## Documentation

### API Documentation

- Add godoc comments for all exported functions and types
- Include examples in godoc where helpful
- Keep documentation up to date with code changes

### README Updates

Update the README.md when:
- Adding new features
- Changing usage patterns
- Updating installation instructions

## Release Process

Releases are handled by maintainers:

1. Version tags follow semantic versioning (v1.2.3)
2. Release notes are generated automatically from commit messages
3. GitHub Actions handles building and publishing

## Getting Help

- Open an issue for bugs or feature requests
- Start a discussion for questions or ideas
- Check existing issues and discussions first

## Code of Conduct

This project follows the [Go Community Code of Conduct](https://golang.org/conduct). Be respectful and professional in all interactions.

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

## Recognition

All contributors will be recognized in the project. Thank you for helping make Vault Watcher better!