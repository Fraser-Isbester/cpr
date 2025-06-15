# cpr - Create Pull Request CLI

A lightweight command-line tool that creates GitHub pull requests with auto-generated Angular format titles and summaries based on your git diff.

## Features

- **Auto-generated PR titles** in Angular commit format (e.g., `feat: add new feature`, `fix: resolve bug`)
- **Intelligent PR summaries** based on your code changes
- **GitHub SDK integration** for native PR creation
- **Customizable** with flags for title, body, and draft status
- **Smart commit type detection** based on file changes and diff content

## Installation

```bash
go install github.com/fraser-isbester/cpr@latest
```

Or build from source:

```bash
git clone https://github.com/fraser-isbester/cpr.git
cd cpr
make install
```

## Usage

### Basic Usage

Simply run `cpr` from your feature branch:

```bash
cpr
```

This will:
1. Analyze the diff between your current branch and the default branch
2. Generate an Angular-format PR title
3. Create a comprehensive PR summary
4. Open a pull request on GitHub

### Command Flags

```bash
cpr [flags]
```

| Flag | Short | Description |
|------|-------|-------------|
| `--title` | `-t` | Custom PR title (overrides auto-generation) |
| `--body` | `-b` | Custom PR body (overrides auto-generation) |
| `--draft` | `-d` | Create PR as draft |
| `--verbose` | `-v` | Enable verbose output |

### Examples

Create a PR with auto-generated title and summary:
```bash
cpr
```

Create a draft PR:
```bash
cpr --draft
```

Create a PR with custom title:
```bash
cpr --title "feat(auth): implement OAuth2 login"
```

Create a PR with custom body:
```bash
cpr --body "This PR implements the new authentication system using OAuth2."
```

## Authentication

The tool requires a GitHub personal access token. Set it as an environment variable:

```bash
export GITHUB_TOKEN=your_token_here
# or
export GH_TOKEN=your_token_here
```

## Angular Commit Types

The tool automatically detects and uses the following commit types:

- `feat`: New features
- `fix`: Bug fixes
- `docs`: Documentation changes
- `style`: Code style changes
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `test`: Test additions or changes
- `build`: Build system changes
- `ci`: CI configuration changes
- `chore`: Other changes

## Development

### Prerequisites

- Go 1.21 or higher
- Git

### Building

```bash
make build
```

### Testing

The project uses Ginkgo for BDD-style testing:

```bash
make test
```

### Available Make Commands

- `make build` - Build the binary
- `make install` - Install the binary to GOPATH/bin
- `make test` - Run all tests
- `make test-coverage` - Run tests with coverage report
- `make clean` - Clean build artifacts
- `make fmt` - Format code
- `make vet` - Run go vet
- `make deps` - Download and tidy dependencies

## License

MIT