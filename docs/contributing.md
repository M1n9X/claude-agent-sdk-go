# Contributing to Claude Agent SDK for Go

Thank you for your interest in contributing to the Claude Agent SDK for Go! This document outlines the guidelines and processes for contributing to the project.

## Getting Started

### Prerequisites
- Go 1.24+
- Claude Code CLI installed: `npm install -g @anthropic-ai/claude-code`
- Valid `CLAUDE_API_KEY` environment variable

### Setup
1. Fork the repository
2. Clone your fork: `git clone https://github.com/your-username/claude-agent-sdk-go`
3. Create a feature branch: `git checkout -b feature/your-feature-name`
4. Run initial tests: `make test`

## Development Workflow

### 1. Code Standards
- Follow Go naming conventions (PascalCase for exports, camelCase for private)
- Write comprehensive comments for exported functions/types
- Use descriptive error messages
- Handle context cancellation properly
- Implement proper error wrapping

### 2. Testing
- Add unit tests for new functionality
- Follow table-driven test patterns
- Aim for >80% coverage on critical paths
- Test both success and error cases

### 3. Commit Messages
Use clear, descriptive commit messages following the format:
```
feat: Add new message type support

- Implement new message types
- Add corresponding tests
- Update documentation
```

### 4. Pull Requests
- Keep PRs focused on a single feature or bug fix
- Include tests for new functionality
- Update documentation as needed
- Reference related issues if applicable

## Code Structure

### Directory Layout
```
├── client.go          # Public Client type
├── query.go          # Public Query function
├── types/            # Public type definitions
├── internal/         # Private implementation
├── examples/         # Example applications
└── tests/            # Test files
```

### Public API Guidelines
- Maintain backward compatibility
- Follow idiomatic Go patterns
- Use context for cancellation
- Implement proper error handling

## Testing Strategy

### Unit Tests
- Fast, isolated tests for individual functions
- No external dependencies
- Located in same directory as source file

### Integration Tests
- Test end-to-end functionality
- May use Claude CLI (with proper setup)
- Located in `tests/` directory

## Code Review Process

1. Submit PR with clear description
2. Ensure all tests pass
3. Address review feedback
4. Get approval from maintainers
5. Merge to main branch

## Reporting Issues

When reporting issues, please include:
- Go version
- Claude CLI version
- Steps to reproduce
- Expected vs actual behavior
- Any relevant error messages

## Development Commands

```bash
make build      # Build the SDK
make test       # Run unit tests
make test-all   # Run all tests including integration
make lint       # Run linters
make fmt        # Format code
make coverage   # Generate coverage report
make examples   # Build all examples
```

## Questions?

If you have questions about contributing, feel free to open an issue or contact the maintainers.

Thank you for contributing to the Claude Agent SDK for Go!