# Claude Agent SDK for Go

A Go SDK for interacting with Claude Code CLI, providing a robust interface for building AI-powered applications using Anthropic's Claude models.

## Overview

The Claude Agent SDK for Go enables Go developers to easily integrate Claude AI capabilities into their applications. The SDK provides:

- **Simple Query Interface**: Execute one-off Claude queries with the `Query` function
- **Rich Message Types**: Handle different message types (user, assistant, system, results)
- **Tool Integration**: Support for various tools (Bash, Read, Write, Edit, Grep, Glob, etc.)
- **Permission Management**: Fine-grained control over tool permissions
- **Agent Definitions**: Create custom agents with specific prompts, tools, and models
- **MCP Server Support**: Integration with Model Context Protocol servers
- **Hook System**: React to lifecycle events (tool use, prompts, etc.)
- **Streaming Support**: Real-time message streaming with partial updates

## Installation

```bash
go get github.com/M1n9X/claude-agent-sdk-go
```

## Prerequisites

- Claude Code CLI must be installed: `npm install -g @anthropic-ai/claude-code`
- Anthropic API key must be set in environment: `ANTHROPIC_API_KEY`

## Quick Start

Here's a simple example to get started:

```go
package main

import (
    "context"
    "fmt"
    "log"

    claude "github.com/M1n9X/claude-agent-sdk-go"
    "github.com/M1n9X/claude-agent-sdk-go/types"
)

func main() {
    ctx := context.Background()
    opts := types.NewClaudeAgentOptions().WithModel("claude-sonnet-4-5-20250929")

    messages, err := claude.Query(ctx, "What is 2+2?", opts)
    if err != nil {
        if types.IsCLINotFoundError(err) {
            log.Fatal("Claude CLI not installed")
        }
        log.Fatal(err)
    }

    for msg := range messages {
        switch m := msg.(type) {
        case *types.AssistantMessage:
            for _, block := range m.Content {
                if tb, ok := block.(*types.TextBlock); ok {
                    fmt.Println(tb.Text)
                }
            }
        case *types.ResultMessage:
            if m.TotalCostUSD != nil {
                fmt.Printf("Cost: $%.4f\n", *m.TotalCostUSD)
            }
        }
    }
}
```

## Features

### Configuration Options

The SDK provides extensive configuration through `ClaudeAgentOptions`:

```go
opts := types.NewClaudeAgentOptions().
    WithModel("claude-sonnet-4-5-20250929").
    WithFallbackModel("claude-3-5-haiku-latest").
    WithAllowedTools("Bash", "Write", "Read").
    WithPermissionMode(types.PermissionModeAcceptEdits).
    WithMaxBudgetUSD(1.0).
    WithSystemPromptString("You are a helpful coding assistant.").
    WithCWD("/path/to/working/directory")
```

By default the Go SDK sends an empty system prompt to the Claude CLI, matching the Python SDK behavior. Use `WithSystemPromptPreset(types.SystemPromptPreset{Type: "preset", Preset: "claude_code"})` to opt into the Claude Code preset (optionally setting `Append` to add extra guidance), or `WithSystemPromptString` to supply your own instructions.

### Structured Outputs

Request validated JSON that matches your schema using `WithJSONSchemaOutput`. The parsed payload is available on `ResultMessage.StructuredOutput`.

```go
schema := map[string]interface{}{
    "type": "object",
    "properties": map[string]interface{}{
        "answer": map[string]interface{}{"type": "string"},
    },
    "required": []interface{}{"answer"},
}

opts := types.NewClaudeAgentOptions().
    WithJSONSchemaOutput(schema)

for msg := range claude.Query(ctx, "Return the answer as JSON", opts) {
    if res, ok := msg.(*types.ResultMessage); ok {
        fmt.Printf("Structured output: %#v\n", res.StructuredOutput)
    }
}
```

### File Checkpointing & Rewind

Enable checkpointing to roll back filesystem changes to any user message UUID.

```go
opts := types.NewClaudeAgentOptions().
    WithEnableFileCheckpointing(true)

client, _ := claude.NewClient(ctx, opts)
defer client.Close(ctx)
_ = client.Connect(ctx)
_ = client.Query(ctx, "Modify files safely")

var checkpoint string
for msg := range client.ReceiveResponse(ctx) {
    if user, ok := msg.(*types.UserMessage); ok && user.UUID != nil {
        checkpoint = *user.UUID
    }
}

// Rewind to the captured checkpoint
_ = client.RewindFiles(ctx, checkpoint)
```

### Base Tools & Betas

Control the default toolset or disable built-ins entirely:

```go
opts := types.NewClaudeAgentOptions().
    WithTools("Read", "Edit").           // or WithToolsPreset(types.ToolsPreset{Type: "preset", Preset: "claude_code"})
    WithBetas(types.SdkBetaContext1M)    // enable extended-context beta
```

### Agent Definitions

Create custom agents with specific capabilities:

```go
model := "sonnet"
agentDef := types.AgentDefinition{
    Description: "Reviews code for best practices and potential issues",
    Prompt:      "You are a code reviewer. Analyze code for bugs, performance issues, security vulnerabilities, and adherence to best practices.",
    Tools:       []string{"Read", "Grep"},
    Model:       &model,
}

opts := types.NewClaudeAgentOptions().
    WithAgent("code-reviewer", agentDef)

messages, err := claude.Query(ctx, "Use the code-reviewer agent to review this code", opts)
```

### Tool Permissions

Control which tools Claude can use:

- `PermissionModeDefault`: Ask user for each tool use
- `PermissionModeAcceptEdits`: Auto-allow file edits
- `PermissionModePlan`: Plan mode (review before execution)
- `PermissionModeBypassPermissions`: Allow all tools (use with caution)

```go
// Custom permission callback
opts := types.NewClaudeAgentOptions().
    WithCanUseTool(func(ctx context.Context, toolName string, input map[string]interface{}, permCtx types.ToolPermissionContext) (interface{}, error) {
        // Implement custom permission logic
        return &types.PermissionResultAllow{Behavior: "allow"}, nil
    })
```

### Hook System

React to various events in the Claude lifecycle:

```go
opts := types.NewClaudeAgentOptions().
    WithHook(types.HookEventPreToolUse, types.HookMatcher{
        Hooks: []types.HookCallbackFunc{
            func(ctx context.Context, input interface{}, toolUseID *string, hookCtx types.HookContext) (interface{}, error) {
                preToolInput := input.(*types.PreToolUseHookInput)
                log.Printf("Tool %s about to execute", preToolInput.ToolName)
                return &types.SyncHookJSONOutput{}, nil
            },
        },
    })
```

### MCP Server Integration

Configure external Model Context Protocol servers:

```go
mcpServers := map[string]interface{}{
    "my-server": map[string]interface{}{
        "type":    "stdio",
        "command": "/path/to/server",
        "args":    []string{"--arg", "value"},
    },
}

opts := types.NewClaudeAgentOptions().
    WithMcpServers(mcpServers)
```

## Error Handling

The SDK provides typed errors for specific failure scenarios:

- `CLINotFoundError`: Claude Code CLI binary not found
- `CLIConnectionError`: Failed to connect to CLI process
- `ProcessError`: CLI subprocess errors (exit codes, crashes)
- `CLIJSONDecodeError`: Invalid JSON from CLI
- `MessageParseError`: Valid JSON but invalid message structure
- `ControlProtocolError`: Control protocol violations
- `PermissionDeniedError`: Permission request denied

```go
messages, err := claude.Query(ctx, "Hello", opts)
if err != nil {
    if types.IsCLINotFoundError(err) {
        log.Fatal("Please install Claude Code CLI: npm install -g @anthropic-ai/claude-code")
    }
    if types.IsCLIConnectionError(err) {
        log.Printf("Connection error: %v", err)
    }
    log.Fatal(err)
}
```

## Advanced Configuration

### Session Management

```go
opts := types.NewClaudeAgentOptions().
    WithResume("session-id").  // Resume an existing session
    WithContinueConversation(true).  // Continue conversation in the same session
    WithForkSession(true)  // Fork the current session
```

### Budget and Limits

```go
opts := types.NewClaudeAgentOptions().
    WithMaxBudgetUSD(5.0).        // Maximum budget for this query
    WithMaxTurns(10).             // Maximum number of turns
    WithMaxThinkingTokens(4096)   // Maximum tokens for internal reasoning
```

### Environment and Extra Arguments

```go
opts := types.NewClaudeAgentOptions().
    WithEnv(map[string]string{
        "CUSTOM_VAR": "value",
    }).
    WithExtraArg("--custom-flag", &someValue)
```

## Message Types

The SDK handles various message types:

- `UserMessage`: Messages from the user to Claude
- `AssistantMessage`: Claude's responses with content blocks
- `SystemMessage`: System notifications and metadata
- `ResultMessage`: Final result with cost/usage info
- `StreamEvent`: Partial message updates during streaming

Content blocks include:

- `TextBlock`: Plain text content
- `ThinkingBlock`: Claude's internal reasoning
- `ToolUseBlock`: Tool invocation requests
- `ToolResultBlock`: Results from tool execution

## Security Considerations

- Use appropriate permission modes based on your use case
- When using `PermissionModeBypassPermissions`, ensure execution in a sandboxed environment
- Set budget limits to control API costs
- Validate and sanitize all inputs before sending to Claude

## Examples

Check out the [examples directory](examples/) for detailed usage examples:

- [Agent examples](examples/advanced/agents/main.go) - Using custom agents with specific tools and prompts
- Additional examples coming soon

## Contributing

Contributions are welcome! Please open an issue or submit a pull request for bug fixes or new features.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

If you encounter issues or have questions, please file an issue on the [GitHub repository](https://github.com/M1n9X/claude-agent-sdk-go).

## Acknowledgment

> This project is based on the original work at [claude-agent-sdk-go](https://github.com/schlunsen/claude-agent-sdk-go).
