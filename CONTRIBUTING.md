# Contributing to RCDS_GO

Thank you for your interest in contributing to RCDS_GO! This document provides guidelines for contributing to the project.

## Code of Conduct

This project adheres to a code of conduct. By participating, you are expected to uphold this code. Please report unacceptable behavior to the project maintainers.

## Getting Started

### Prerequisites

- Go 1.21 or later
- Git
- Make

### Setting Up Your Development Environment

1. Fork the repository on GitHub
2. Clone your fork locally:
   ```bash
   git clone https://github.com/YOUR-USERNAME/RCDS_GO.git
   cd RCDS_GO
   ```

3. Add the upstream repository:
   ```bash
   git remote add upstream https://github.com/String-Reconciliation-Ditributed-System/RCDS_GO.git
   ```

4. Install dependencies:
   ```bash
   go mod download
   ```

5. Build the project:
   ```bash
   make build
   ```

6. Run tests:
   ```bash
   make test
   ```

## Development Workflow

### Creating a Branch

Create a new branch for your work:

```bash
git checkout -b feature/your-feature-name
```

Use descriptive branch names:
- `feature/` for new features
- `bugfix/` for bug fixes
- `docs/` for documentation changes
- `test/` for test improvements

### Making Changes

1. Make your changes in the appropriate files
2. Add tests for new functionality
3. Ensure all tests pass: `make test`
4. Format your code: `make fmt`
5. Run linters: `make lint`
6. Run vet: `make vet`

### Commit Messages

Write clear commit messages that describe your changes:

```
Short (50 chars or less) summary

More detailed explanatory text, if necessary. Wrap it to about 72
characters. The blank line separating the summary from the body is
critical.

- Bullet points are okay
- Use present tense ("Add feature" not "Added feature")
- Reference issues and pull requests liberally
```

### Testing

- Write unit tests for all new code
- Ensure all tests pass before submitting
- Aim for high test coverage
- Run tests with: `make test`
- Check coverage with: `make test-coverage`

### Code Style

- Follow standard Go conventions
- Use `gofmt` to format code
- Run `go vet` to catch common errors
- Use `golangci-lint` for additional linting
- Write clear, concise comments for exported functions
- Use meaningful variable and function names

### Submitting a Pull Request

1. Push your changes to your fork:
   ```bash
   git push origin feature/your-feature-name
   ```

2. Open a pull request on GitHub
3. Fill in the pull request template
4. Link any related issues
5. Wait for review

### Pull Request Guidelines

- Keep pull requests focused on a single feature or fix
- Include tests for new functionality
- Update documentation as needed
- Ensure CI/CD checks pass
- Respond to review comments promptly

## Project Structure

```
RCDS_GO/
├── cmd/              # Command-line application
├── pkg/              # Library code
│   ├── file/         # File utilities
│   ├── lib/          # Core libraries
│   │   ├── algorithm/  # Reconciliation algorithms
│   │   └── genSync/    # Generic sync interface
│   ├── set/          # Set data structure
│   └── util/         # Utility functions
├── docs/             # Documentation
└── .github/          # GitHub workflows
```

## Running Tests

### Unit Tests

```bash
make test
```

### Test Coverage

```bash
make test-coverage
```

This generates a coverage report in `coverage.html`.

### Linting

```bash
make lint
```

## Documentation

- Update documentation for any user-facing changes
- Add godoc comments for exported functions
- Update README.md for major changes
- Add examples for new features

## Reporting Issues

### Bug Reports

When reporting bugs, include:

- Go version
- Operating system
- Steps to reproduce
- Expected behavior
- Actual behavior
- Error messages or logs

### Feature Requests

When requesting features, include:

- Use case description
- Proposed solution
- Alternative solutions considered
- Potential impact

## Questions?

If you have questions, please:

1. Check existing documentation
2. Search existing issues
3. Open a new issue with the question label

## License

By contributing to RCDS_GO, you agree that your contributions will be licensed under the same license as the project.
