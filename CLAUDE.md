# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`cpr` is a lightweight CLI tool that automates GitHub pull request creation with intelligent title and summary generation based on git diffs. It follows the Angular commit convention and uses clean architecture principles.

## Development Commands

### Building and Installation
```bash
make build          # Build the binary
make install        # Install to GOPATH/bin
make all           # Clean, deps, fmt, vet, test, and build
```

### Testing
```bash
make test          # Run all tests with Ginkgo
make test-coverage # Run tests with coverage report
```

### Code Quality
```bash
make fmt           # Format code
make vet           # Run go vet
make lint          # Run golangci-lint (requires golangci-lint installed)
```

### Dependency Management
```bash
make deps          # Download and tidy dependencies
```

### Running a Single Test
```bash
# Run specific test suite
go test -v ./internal/commit -ginkgo.focus="should detect feat type"

# Run tests for a specific package
go test -v ./internal/commit
```

## Architecture

The codebase follows clean architecture with clear separation of concerns:

- **`/cmd`**: CLI layer using Cobra framework. The main entry point (`root.go`) orchestrates the PR creation flow.
- **`/internal/commit`**: Core business logic for analyzing git diffs and generating Angular-format titles. Uses pattern matching and diff analysis to determine commit types.
- **`/internal/git`**: Git operations wrapper that handles repository interactions, branch detection, and diff generation.
- **`/internal/github`**: GitHub API client using google/go-github for creating pull requests.

## Key Implementation Details

### Commit Type Detection (`internal/commit/analyzer.go`)
The analyzer determines commit types through:
1. File pattern analysis (e.g., test files → "test", docs → "docs")
2. Diff content analysis (bug fixes, performance improvements)
3. Code structure changes (new functions, types, imports)

### Testing Approach
- Uses Ginkgo BDD framework with Gomega matchers
- Tests are colocated with implementation files
- Focus on behavior-driven test descriptions

### Authentication
The tool uses GitHub tokens from environment variables:
- Primary: `GITHUB_TOKEN`
- Fallback: `GH_TOKEN`

### Angular Commit Types Supported
feat, fix, docs, style, refactor, perf, test, build, ci, chore

## Common Development Tasks

When adding new features:
1. Follow the existing package structure
2. Add appropriate tests using Ginkgo's BDD style
3. Update the commit analyzer if adding new commit type detection patterns
4. Ensure cross-platform compatibility (tool supports Darwin, Linux, Windows)