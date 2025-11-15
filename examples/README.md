# Examples

This directory contains various examples demonstrating how to use the Claude Agent SDK. The examples are organized into different categories based on their complexity and use cases.

## Basic Examples

### quick_start

- **Functionality**: Demonstrates a quick start with three basic usage patterns:
  - Basic query with default options
  - Query with custom options (system prompt, max turns)
  - Query with tools enabled (Read, Write operations)
- **Run**: `cd examples/basic/quick_start && go run main.go`

### simple_query

- **Functionality**: Shows the simplest way to use the SDK - send a prompt and get responses using the one-shot Query function.
- **Run**: `cd examples/basic/simple_query && go run main.go`

### interactive_client

- **Functionality**: Demonstrates an interactive client for multi-turn conversations that maintains session state between exchanges.
- **Run**: `cd examples/basic/interactive_client && go run main.go`

## Advanced Examples

### agents

- **Functionality**: Shows how to use custom agents with specific tools, prompts, and models:
  - Code reviewer agent for analyzing code quality
  - Documentation writer agent for generating docs
  - Multiple agents example for different tasks
- **Run**: `cd examples/advanced/agents && go run main.go`

### plugin_example

- **Functionality**: Demonstrates how to build and use custom plugins with the Claude Agent SDK.
- **Run**: `cd examples/advanced/plugin_example && go run main.go`

## Configuration Examples

### configurable_channels

- **Functionality**: Shows how to configure custom channels for different types of communication with Claude.
- **Run**: `cd examples/configuration/configurable_channels && go run main.go`

### max_budget_usd

- **Functionality**: Demonstrates setting a maximum budget in USD to control API costs.
- **Run**: `cd examples/configuration/max_budget_usd && go run main.go`

### setting_sources

- **Functionality**: Shows how to use different setting sources (user, project) for configuration.
- **Run**: `cd examples/configuration/setting_sources && go run main.go`

### system_prompt

- **Functionality**: Demonstrates customizing Claude's behavior with system prompts.
- **Run**: `cd examples/configuration/system_prompt && go run main.go`

### type_safe_accessors

- **Functionality**: Shows type-safe accessors for configuration values.
- **Run**: `cd examples/configuration/type_safe_accessors && go run main.go`

## Hooks Examples

### with_hooks

- **Functionality**: Demonstrates how to use hooks for intercepting and modifying requests/responses.
- **Run**: `cd examples/hooks/with_hooks && go run main.go`

## MCP (Model Context Protocol) Examples

### mcp_calculator

- **Functionality**: Shows how to create an MCP server with a calculator tool that Claude can call.
- **Run**: `cd examples/mcp/mcp_calculator && go run main.go`

### streaming_mode

- **Functionality**: Demonstrates various patterns for building applications with the Claude SDK streaming interface:
  - Basic streaming
  - Multi-turn conversations
  - Concurrent send/receive
  - Interrupt handling
  - Custom options
- **Run**: `cd examples/mcp/streaming_mode && go run main.go`

### streaming_mode_comprehensive

- **Functionality**: A more comprehensive example of streaming mode with additional patterns.
- **Run**: `cd examples/mcp/streaming_mode_comprehensive && go run main.go`

### streaming_mode_conversation

- **Functionality**: Demonstrates streaming mode specifically for conversation-style interactions.
- **Run**: `cd examples/mcp/streaming_mode_conversation && go run main.go`

## Permissions Examples

### tool_permission_callback

- **Functionality**: Shows how to implement custom callbacks for handling tool permission requests.
- **Run**: `cd examples/permissions/tool_permission_callback && go run main.go`

### with_permissions

- **Functionality**: Demonstrates how to set up and use various permission levels for tools.
- **Run**: `cd examples/permissions/with_permissions && go run main.go`

## Utilities Examples

### include_partial_messages

- **Functionality**: Shows how to include partial messages in the response stream.
- **Run**: `cd examples/utilities/include_partial_messages && go run main.go`

### stderr_callback

- **Functionality**: Demonstrates how to use callbacks to capture and handle stderr output.
- **Run**: `cd examples/utilities/stderr_callback && go run main.go`

## Prerequisites

Before running any of these examples, ensure you have:

1. Go 1.21 or later installed
2. Valid Anthropic API key set as an environment variable:

   ```bash
   export ANTHROPIC_API_KEY=your_api_key_here
   ```

## Running Examples

To run any example, navigate to the example's directory and run:

```bash
go run main.go
```

Most examples will demonstrate their functionality directly and may prompt for input where applicable.
