# Contributing to Sentinel-Remediator

Thank you for your interest in contributing! This document provides guidelines and information about contributing to this project.

## Development Setup

### Prerequisites

- Go 1.21+
- Node.js 20+
- Docker & Docker Compose
- Git

### Getting Started

```bash
# Clone the repository
git clone https://github.com/yourusername/sentinel-remediator.git
cd sentinel-remediator

# Install dependencies
make deps
cd dashboard && npm install && cd ..

# Copy environment file
cp .env.example .env
# Edit .env with your API keys

# Run in development mode
make dev
```

## Code Style

### Go
- Follow [Effective Go](https://golang.org/doc/effective_go)
- Run `golangci-lint run ./...` before committing
- Write tests for new features

### TypeScript/React
- Use functional components with hooks
- Run `npm run lint` before committing

## Pull Request Process

1. Create a feature branch from `develop`
2. Make your changes with clear commit messages
3. Ensure all tests pass: `make test`
4. Update documentation if needed
5. Submit PR against `develop` branch

## Commit Message Format

```
type(scope): description

[optional body]
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`

Example: `feat(agent): add retry logic for build failures`

## Architecture Decisions

See [docs/architecture.md](docs/architecture.md) for architectural decisions and patterns used.

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
