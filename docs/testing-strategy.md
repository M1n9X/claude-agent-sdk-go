# Testing Strategy

The Claude Agent SDK for Go follows a comprehensive testing approach to ensure reliability and correctness across different components.

## Test Categories

### Unit Tests
- Test individual functions and methods in isolation
- Fast execution with no external dependencies
- Located alongside source files (e.g., `messages_test.go` with `messages.go`)
- Focus on business logic, data validation, and error handling

### Integration Tests
- Test end-to-end functionality with real Claude CLI
- Located in `tests/integration/` directory
- Require `CLAUDE_API_KEY` environment variable
- Validate complete workflows and interactions

### Performance Tests
- Located in `tests/performance/` directory
- Measure response times and resource usage
- Identify bottlenecks and optimization opportunities

### MCP Tests
- Located in `tests/mcp/` directory
- Test Model Context Protocol functionality
- Validate MCP server interactions

## Test Organization

### Unit Tests Structure
```
types/
├── messages.go
├── messages_test.go
├── errors.go
└── errors_test.go

internal/
├── parser/
│   ├── message_parser.go
│   └── message_parser_test.go
├── protocol/
│   ├── query.go
│   └── query_test.go
└── transport/
    ├── transport.go
    └── transport_test.go
```

### Integration Tests Structure
```
tests/
├── integration/
├── mcp/
└── performance/
```

## Test Guidelines

### Writing Tests
- Use table-driven tests for multiple scenarios
- Include both success and failure cases
- Test edge cases and error conditions
- Use descriptive test names

### Example Test Pattern
```go
func TestFunctionName(t *testing.T) {
    tests := []struct {
        name     string
        input    interface{}
        expected interface{}
        wantErr  bool
    }{
        {
            name:     "valid input returns expected result",
            input:    validInput,
            expected: expectedResult,
            wantErr:  false,
        },
        {
            name:     "invalid input returns error",
            input:    invalidInput,
            expected: nil,
            wantErr:  true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

## Test Execution

### Running Tests
```bash
# Run all unit tests
make test

# Run all tests (including integration)
make test-all

# Run specific test package
go test -v ./types/...

# Run with coverage
go test -cover ./...
```

### Test Coverage
- Target >80% coverage on critical paths
- Focus on functionality over line coverage
- Prioritize testing error paths and edge cases

## Mocking Strategy

- Use real implementations when possible
- Create interfaces for external dependencies
- Use mocks for external services and complex dependencies
- Avoid over-mocking business logic

## Continuous Integration

Tests run automatically in CI:
- Unit tests on every commit
- Integration tests on PRs and merges
- Performance tests on schedule
- Linting and formatting checks