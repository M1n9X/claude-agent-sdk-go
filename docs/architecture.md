# Architecture Overview

The Claude Agent SDK for Go follows a layered architecture pattern with clear separation of concerns:

## System Architecture

```
┌─────────────────────────────────────────┐
│           User Application              │
├─────────────────────────────────────────┤
│         Public API Layer                │
│    (Query, Client)                      │
├─────────────────────────────────────────┤
│        Internal Orchestration           │
│    (Client implementation)              │
├─────────────────────────────────────────┤
│         Control Protocol                │
│    (Permissions, Hooks, MCP)            │
├─────────────────────────────────────────┤
│        Message Parser                   │
│    (JSON to Go types conversion)        │
├─────────────────────────────────────────┤
│          Transport Layer                │
│    (SubprocessCLI, HTTP, MCP)          │
├─────────────────────────────────────────┤
│       Claude Code CLI (Node.js)         │
└─────────────────────────────────────────┘
```

## Key Components

### Public API Layer
- `Query()` function for one-shot interactions
- `Client` type for interactive, bidirectional conversations
- `ClaudeAgentOptions` builder for configuration

### Internal Packages
- `internal/transport` - Handles subprocess communication
- `internal/parser` - Converts JSON responses to Go types
- `internal/protocol` - Implements control protocol (permissions, hooks)
- `internal/mcp` - Model Context Protocol implementation
- `internal/logging` - Centralized logging utilities

### Type Definitions
- `types/messages.go` - Message types and content blocks
- `types/control.go` - Control protocol types
- `types/errors.go` - Error definitions
- `types/options.go` - Configuration options

## Data Flow

1. User calls `Query()` or `Client.Query()`
2. Options are validated and prepared
3. Transport layer establishes connection to Claude CLI
4. Messages are sent and received via JSON lines protocol
5. Response JSON is parsed into Go types
6. Control protocol handles permissions and hooks
7. Messages are returned to the user

## Concurrency Model

The SDK uses Go's concurrency patterns:
- Channels for message streaming
- Goroutines for background processing
- Context for cancellation and timeouts
- Mutexes for shared state protection