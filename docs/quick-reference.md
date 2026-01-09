# Quick Reference Guide

A concise reference for common Claude Agent SDK operations.

## Table of Contents

- [Basic Usage](#basic-usage)
- [Configuration](#configuration)
- [Custom Tools](#custom-tools)
- [Hooks](#hooks)
- [Permissions](#permissions)
- [Error Handling](#error-handling)
- [Message Handling](#message-handling)

## Basic Usage

### Simple Query
```go
ctx := context.Background()
opts := types.NewClaudeAgentOptions().WithModel("claude-sonnet-4-5-20250929")

messages, err := claude.Query(ctx, "Hello, Claude!", opts)
if err != nil {
    log.Fatal(err)
}

for msg := range messages {
    if assistantMsg, ok := msg.(*types.AssistantMessage); ok {
        for _, block := range assistantMsg.Content {
            if textBlock, ok := block.(*types.TextBlock); ok {
                fmt.Println(textBlock.Text)
            }
        }
    }
}
```

### Interactive Client
```go
client, _ := claude.NewClient(ctx, opts)
defer client.Close(ctx)

client.Connect(ctx)
client.Query(ctx, "First question")
for msg := range client.ReceiveResponse(ctx) {
    // Process messages
}

client.Query(ctx, "Second question")
for msg := range client.ReceiveResponse(ctx) {
    // Process messages
}
```

## Configuration

### Common Options
```go
opts := types.NewClaudeAgentOptions().
    WithModel("claude-sonnet-4-5-20250929").
    WithFallbackModel("claude-3-5-haiku-latest").
    WithMaxTurns(10).
    WithMaxBudgetUSD(1.0).
    WithSystemPromptString("You are a helpful assistant.").
    WithAllowedTools("Bash", "Read", "Write").
    WithPermissionMode(types.PermissionModeAcceptEdits)
```

### Session Management
```go
// Resume existing session
opts := types.NewClaudeAgentOptions().WithResume("session-id")

// Continue conversation
opts := types.NewClaudeAgentOptions().WithContinueConversation(true)

// Fork session
opts := types.NewClaudeAgentOptions().WithForkSession(true)
```

### Structured Output
```go
schema := map[string]interface{}{
    "type": "object",
    "properties": map[string]interface{}{
        "answer": map[string]interface{}{"type": "string"},
        "confidence": map[string]interface{}{"type": "number"},
    },
    "required": []interface{}{"answer"},
}

opts := types.NewClaudeAgentOptions().WithJSONSchemaOutput(schema)

for msg := range claude.Query(ctx, "Answer as JSON", opts) {
    if result, ok := msg.(*types.ResultMessage); ok {
        fmt.Printf("Structured: %+v\n", result.StructuredOutput)
    }
}
```

## Custom Tools

### Method 1: SimpleTool (Python @tool style)
```go
tool := types.SimpleTool{
    Name:        "greet",
    Description: "Greet a user",
    Parameters: map[string]types.SimpleParam{
        "name": {Type: "string", Description: "User's name", Required: true},
    },
    Handler: func(ctx context.Context, args map[string]interface{}) (*types.ToolResult, error) {
        name := args["name"].(string)
        return types.NewMcpToolResult(
            types.TextBlock{Type: "text", Text: fmt.Sprintf("Hello, %s!", name)},
        ), nil
    },
}
mcpTool, _ := tool.Build()
```

### Method 2: Fluent API
```go
tool, _ := types.Tool("greet", "Greet a user").
    Param("name", "string", "User's name", true).
    Handle(func(ctx context.Context, args map[string]interface{}) (*types.ToolResult, error) {
        name := args["name"].(string)
        return types.NewMcpToolResult(
            types.TextBlock{Type: "text", Text: fmt.Sprintf("Hello, %s!", name)},
        ), nil
    })
```

### Method 3: QuickTool
```go
tool, _ := types.QuickTool("greet", "Greet a user",
    map[string]string{"name": "string"},
    func(ctx context.Context, args map[string]interface{}) (*types.ToolResult, error) {
        name := args["name"].(string)
        return types.NewMcpToolResult(
            types.TextBlock{Type: "text", Text: fmt.Sprintf("Hello, %s!", name)},
        ), nil
    },
)
```

### Using Custom Tools
```go
server := types.CreateToolServer("my-tools", "1.0.0", []types.McpTool{tool})

opts := types.NewClaudeAgentOptions().
    WithMcpServers(map[string]interface{}{"tools": server}).
    WithAllowedTools("mcp__tools__greet")

messages, _ := claude.Query(ctx, "Greet Alice", opts)
```

## Hooks

### All Hook Events
```go
opts := types.NewClaudeAgentOptions().
    // Before tool execution
    WithHook(types.HookEventPreToolUse, types.HookMatcher{
        Hooks: []types.HookCallbackFunc{preToolHook},
    }).
    // After tool execution
    WithHook(types.HookEventPostToolUse, types.HookMatcher{
        Hooks: []types.HookCallbackFunc{postToolHook},
    }).
    // Before model call
    WithHook(types.HookEventPrePrompt, types.HookMatcher{
        Hooks: []types.HookCallbackFunc{prePromptHook},
    }).
    // After model response
    WithHook(types.HookEventPostPrompt, types.HookMatcher{
        Hooks: []types.HookCallbackFunc{postPromptHook},
    }).
    // Error handling
    WithHook(types.HookEventOnError, types.HookMatcher{
        Hooks: []types.HookCallbackFunc{errorHook},
    })
```

### Hook Implementation
```go
func preToolHook(ctx context.Context, input interface{}, toolUseID *string, hookCtx types.HookContext) (interface{}, error) {
    if inputMap, ok := input.(map[string]interface{}); ok {
        toolName := inputMap["tool_name"].(string)
        log.Printf("Tool %s about to execute", toolName)
    }
    return &types.SyncHookJSONOutput{}, nil
}
```

### Hook with Matcher (Regex)
```go
bashMatcher := "Bash"
opts := types.NewClaudeAgentOptions().
    WithHook(types.HookEventPreToolUse, types.HookMatcher{
        Matcher: &bashMatcher, // Only match Bash tool
        Hooks:   []types.HookCallbackFunc{bashHook},
    })
```

## Permissions

### Permission Modes
```go
// Ask for each tool
opts := types.NewClaudeAgentOptions().
    WithPermissionMode(types.PermissionModeDefault)

// Auto-allow file edits
opts := types.NewClaudeAgentOptions().
    WithPermissionMode(types.PermissionModeAcceptEdits)

// Plan mode
opts := types.NewClaudeAgentOptions().
    WithPermissionMode(types.PermissionModePlan)

// Allow all (use with caution!)
opts := types.NewClaudeAgentOptions().
    WithPermissionMode(types.PermissionModeBypassPermissions)
```

### Custom Permission Callback
```go
canUseTool := func(ctx context.Context, toolName string, input map[string]interface{}, permCtx types.ToolPermissionContext) (interface{}, error) {
    // Allow read-only tools
    if toolName == "Read" || toolName == "Grep" {
        return &types.PermissionResultAllow{Behavior: "allow"}, nil
    }
    
    // Deny dangerous operations
    if toolName == "Bash" {
        command := input["command"].(string)
        if strings.Contains(command, "rm -rf") {
            return &types.PermissionResultDeny{
                Behavior: "deny",
                Message:  "Dangerous command blocked",
            }, nil
        }
    }
    
    // Modify tool input
    if toolName == "Write" {
        modifiedInput := make(map[string]interface{})
        for k, v := range input {
            modifiedInput[k] = v
        }
        modifiedInput["file_path"] = "/safe/" + input["file_path"].(string)
        
        return &types.PermissionResultAllow{
            Behavior:     "allow",
            UpdatedInput: &modifiedInput,
        }, nil
    }
    
    return &types.PermissionResultAllow{Behavior: "allow"}, nil
}

opts := types.NewClaudeAgentOptions().WithCanUseTool(canUseTool)
```

## Error Handling

### Type Guards
```go
messages, err := claude.Query(ctx, "Hello", opts)
if err != nil {
    if types.IsCLINotFoundError(err) {
        log.Fatal("Please install Claude CLI: npm install -g @anthropic-ai/claude-code")
    }
    if types.IsCLIConnectionError(err) {
        log.Fatal("Failed to connect to Claude CLI")
    }
    if types.IsSessionNotFoundError(err) {
        log.Fatal("Session not found")
    }
    log.Fatal(err)
}
```

### Error Types
- `CLINotFoundError` - CLI binary not found
- `CLIConnectionError` - Connection failed
- `ProcessError` - CLI process error
- `CLIJSONDecodeError` - Invalid JSON from CLI
- `MessageParseError` - Invalid message structure
- `ControlProtocolError` - Protocol violation
- `PermissionDeniedError` - Permission denied
- `SessionNotFoundError` - Session not found

## Message Handling

### Type Assertions
```go
for msg := range messages {
    switch m := msg.(type) {
    case *types.UserMessage:
        fmt.Printf("User: %v\n", m.Content)
        
    case *types.AssistantMessage:
        for _, block := range m.Content {
            switch b := block.(type) {
            case *types.TextBlock:
                fmt.Printf("Text: %s\n", b.Text)
            case *types.ThinkingBlock:
                fmt.Printf("Thinking: %s\n", b.Thinking)
            case *types.ToolUseBlock:
                fmt.Printf("Tool: %s(%+v)\n", b.Name, b.Input)
            }
        }
        
    case *types.SystemMessage:
        if m.IsWarning() {
            log.Printf("Warning: %+v\n", m.Data)
        }
        
    case *types.ResultMessage:
        fmt.Printf("Cost: $%.4f\n", *m.TotalCostUSD)
        fmt.Printf("Duration: %dms\n", m.DurationMs)
        fmt.Printf("Turns: %d\n", m.NumTurns)
    }
}
```

### Helper Functions
```go
// Type-safe casting
if assistantMsg, ok := types.AsAssistant(msg); ok {
    // Process assistant message
}

if resultMsg, ok := types.AsResult(msg); ok {
    // Process result
}
```

## File Checkpointing

### Enable and Use
```go
opts := types.NewClaudeAgentOptions().
    WithEnableFileCheckpointing(true)

client, _ := claude.NewClient(ctx, opts)
defer client.Close(ctx)

client.Connect(ctx)
client.Query(ctx, "Modify some files")

var checkpoint string
for msg := range client.ReceiveResponse(ctx) {
    if userMsg, ok := msg.(*types.UserMessage); ok && userMsg.UUID != nil {
        checkpoint = *userMsg.UUID
    }
}

// Rewind to checkpoint
client.RewindFiles(ctx, checkpoint)
```

## MCP Servers

### External Stdio Server
```go
mcpServers := map[string]interface{}{
    "my-server": map[string]interface{}{
        "type":    "stdio",
        "command": "/path/to/server",
        "args":    []string{"--arg", "value"},
        "env": map[string]string{
            "API_KEY": "secret",
        },
    },
}

opts := types.NewClaudeAgentOptions().WithMcpServers(mcpServers)
```

### External SSE Server
```go
mcpServers := map[string]interface{}{
    "sse-server": map[string]interface{}{
        "type": "sse",
        "url":  "https://example.com/sse",
        "headers": map[string]string{
            "Authorization": "Bearer token",
        },
    },
}
```

## Agent Definitions

```go
model := "sonnet"
agentDef := types.AgentDefinition{
    Description: "Code reviewer",
    Prompt:      "You are a code reviewer. Analyze code for bugs and best practices.",
    Tools:       []string{"Read", "Grep"},
    Model:       &model,
}

opts := types.NewClaudeAgentOptions().
    WithAgent("code-reviewer", agentDef)

messages, _ := claude.Query(ctx, "Use the code-reviewer agent", opts)
```

## Best Practices

### 1. Always Clean Up
```go
client, err := claude.NewClient(ctx, opts)
if err != nil {
    log.Fatal(err)
}
defer client.Close(ctx) // Always defer Close()
```

### 2. Use Context for Timeouts
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

messages, err := claude.Query(ctx, "Long task", opts)
```

### 3. Handle All Message Types
```go
for msg := range messages {
    switch msg.(type) {
    case *types.UserMessage:
        // Handle user messages
    case *types.AssistantMessage:
        // Handle assistant messages
    case *types.SystemMessage:
        // Handle system messages
    case *types.ResultMessage:
        // Handle results
    }
}
```

### 4. Validate Tool Inputs
```go
tool := types.SimpleTool{
    Name: "process",
    // ... other fields
    Handler: func(ctx context.Context, args map[string]interface{}) (*types.ToolResult, error) {
        // Validate inputs
        data, ok := args["data"].(string)
        if !ok || data == "" {
            return types.NewErrorMcpToolResult("Invalid data"), nil
        }
        
        // Process data
        result := processData(data)
        return types.NewMcpToolResult(
            types.TextBlock{Type: "text", Text: result},
        ), nil
    },
}
```

### 5. Use Type Guards for Errors
```go
if err != nil {
    switch {
    case types.IsCLINotFoundError(err):
        // Handle CLI not found
    case types.IsCLIConnectionError(err):
        // Handle connection error
    case types.IsSessionNotFoundError(err):
        // Handle session not found
    default:
        // Handle other errors
    }
}
```

## See Also

- [Python SDK Alignment](python-sdk-alignment.md) - Complete feature comparison
- [Feature Checklist](feature-checklist.md) - Implementation status
- [Architecture](architecture.md) - System architecture
- [Examples](../examples/) - Working examples
