# Contributing to KubeCaps

Thank you for your interest in contributing to KubeCaps! This document provides guidelines and instructions for contributing.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Making Changes](#making-changes)
- [Testing](#testing)
- [Submitting Changes](#submitting-changes)
- [Code Style](#code-style)
- [Commit Messages](#commit-messages)

## Code of Conduct

This project adheres to a code of conduct. By participating, you are expected to uphold this code. Please report unacceptable behavior to the project maintainers.

## Getting Started

1. Fork the repository on GitHub
2. Clone your fork locally
3. Create a new branch for your changes
4. Make your changes
5. Push to your fork
6. Submit a pull request

## Development Setup

### Prerequisites

- Go 1.21 or higher
- Access to a Kubernetes cluster (for integration testing)
- Prometheus instance (for testing)
- Make (optional, but recommended)

### Setup Steps

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/kubecaps.git
cd kubecaps

# Add upstream remote
git remote add upstream https://github.com/vasudevchavan/kubecaps.git

# Install dependencies
go mod download

# Build the project
make build

# Run tests
make test
```

## Making Changes

### Branch Naming

Use descriptive branch names:
- `feature/add-new-predictor` - for new features
- `fix/handle-nil-pointer` - for bug fixes
- `docs/update-readme` - for documentation
- `refactor/simplify-engine` - for refactoring
- `test/add-unit-tests` - for adding tests

### Development Workflow

1. **Create a branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes**
   - Write clean, readable code
   - Add tests for new functionality
   - Update documentation as needed

3. **Test your changes**
   ```bash
   make test
   make lint
   ```

4. **Commit your changes**
   ```bash
   git add .
   git commit -m "feat: add new prediction model"
   ```

5. **Keep your branch updated**
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

6. **Push to your fork**
   ```bash
   git push origin feature/your-feature-name
   ```

## Testing

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make coverage

# Run specific package tests
go test -v ./internal/predictor/...

# Run benchmarks
make bench
```

### Writing Tests

- Write table-driven tests when possible
- Test edge cases and error conditions
- Use meaningful test names
- Add benchmarks for performance-critical code

Example:
```go
func TestEWMA(t *testing.T) {
    tests := []struct {
        name     string
        values   []float64
        alpha    float64
        expected float64
    }{
        {
            name:     "empty values",
            values:   []float64{},
            alpha:    0.3,
            expected: 0,
        },
        // More test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := EWMA(tt.values, tt.alpha)
            if result != tt.expected {
                t.Errorf("got %v, want %v", result, tt.expected)
            }
        })
    }
}
```

## Submitting Changes

### Pull Request Process

1. **Update documentation** - Ensure README and other docs reflect your changes
2. **Add tests** - All new code should have tests
3. **Run the full test suite** - `make verify`
4. **Update CHANGELOG** - Add your changes to the unreleased section
5. **Create a pull request** with a clear title and description

### Pull Request Template

```markdown
## Description
Brief description of the changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
Describe the tests you ran

## Checklist
- [ ] My code follows the project's style guidelines
- [ ] I have performed a self-review
- [ ] I have commented my code where necessary
- [ ] I have updated the documentation
- [ ] My changes generate no new warnings
- [ ] I have added tests that prove my fix/feature works
- [ ] New and existing tests pass locally
```

## Code Style

### Go Style Guide

Follow the [Effective Go](https://golang.org/doc/effective_go.html) guidelines and:

- Use `gofmt` to format your code
- Use `goimports` to organize imports
- Follow Go naming conventions
- Keep functions small and focused
- Add comments for exported functions
- Use meaningful variable names

### Code Organization

```
internal/
├── cli/          # CLI commands and flags
├── config/       # Configuration management
├── evaluator/    # Autoscaling evaluation logic
├── k8s/          # Kubernetes client wrappers
├── output/       # Output formatting
├── predictor/    # Prediction algorithms
├── prometheus/   # Prometheus client
└── recommender/  # Recommendation generation
```

### Documentation

- Add godoc comments for all exported functions, types, and constants
- Include examples in documentation where helpful
- Keep comments up-to-date with code changes

Example:
```go
// EWMA calculates the Exponentially Weighted Moving Average of a time series.
// It gives more weight to recent values based on the alpha parameter.
//
// Parameters:
//   - values: The time series data points
//   - alpha: The smoothing factor (0 < alpha <= 1)
//
// Returns the EWMA value, or 0 if values is empty.
func EWMA(values []float64, alpha float64) float64 {
    // Implementation...
}
```

## Commit Messages

Follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

### Examples

```
feat(predictor): add Holt-Winters forecasting model

Implement triple exponential smoothing for better seasonal prediction.
Includes tests and benchmarks.

Closes #123
```

```
fix(evaluator): handle nil HPA config gracefully

Previously crashed when HPA was not configured. Now returns
appropriate default score.

Fixes #456
```

## Questions?

If you have questions or need help:

1. Check existing issues and discussions
2. Open a new issue with the `question` label
3. Reach out to maintainers

Thank you for contributing to KubeCaps! 🎉