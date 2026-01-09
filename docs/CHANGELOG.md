# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added - Complete Python SDK Alignment

This release achieves **100% feature parity** with the official Python SDK.

#### Hook System Enhancements
- Added `HookEventPrePrompt` - Called before sending messages to model
- Added `HookEventPostPrompt` - Called after receiving response from model
- Added `HookEventPreResponse` - Called before sending response to user
- Added `HookEventPostResponse` - Called after sending response to user
- Added `HookEventPostCompact` - Called after context compaction
- Added `HookEventOnError` - Called when an error occurs
- Added corresponding hook input types:
  - `PrePromptHookInput`
  - `PostPromptHookInput`
  - `PreResponseHookInput`
  - `PostResponseHookInput`
  - `PostCompactHookInput`
  - `OnErrorHookInput`
- Added corresponding hook output types:
  - `PrePromptHookSpecificOutput`
  - `PostPromptHookSpecificOutput`
  - `PreResponseHookSpecificOutput`
  - `PostResponseHookSpecificOutput`
  - `PostCompactHookSpecificOutput`
  - `OnErrorHookSpecificOutput`

#### Decorator-Style Tool Creation API
- Added `SimpleTool` struct for Python `@tool` decorator-style tool creation
- Added `SimpleParam` for simplified parameter definitions
- Added `Tool()` function for fluent API tool creation
- Added `QuickTool()` for ultra-concise tool creation
- Added `ToolDecorator` for builder-pattern tool creation
- Added `MustHandle()` and `MustQuickTool()` for panic-on-error variants

#### Examples
- Added `examples/mcp/decorator_style_tools/` - Demonstrates all tool creation methods
- Added `examples/hooks/comprehensive_hooks/` - Demonstrates all 12 hook events

#### Documentation
- Added `docs/python-sdk-alignment.md` - Complete feature parity documentation
- Updated `README.md` with new hook events and tool creation APIs
- Added migration guide from Python SDK to Go SDK
- Added API comparison tables

### Changed
- Enhanced hook system to support all 12 hook events (was 6, now 12)
- Improved tool creation ergonomics with multiple API styles
- Updated examples to showcase new features

### Technical Details
- All hook events now fully supported and tested
- Tool creation APIs provide Python-like ergonomics in Go
- Complete type safety maintained throughout
- Zero breaking changes to existing APIs

## [0.1.0] - 2025-01-09

### Added
- Initial release of Claude Agent SDK for Go
- `Query()` function for one-shot interactions
- `Client` type for interactive, bidirectional conversations
- Support for all Claude Code CLI features:
  - Tool permissions and callbacks
  - MCP server integration (stdio, SSE, HTTP)
  - SDK MCP servers (in-process tools)
  - Session management (resume, fork, continue)
  - File checkpointing and rewind
  - Structured outputs (JSON schema)
  - Agent definitions
  - Hooks (PreToolUse, PostToolUse, UserPromptSubmit, Stop, SubagentStop, PreCompact)
- Comprehensive error handling with typed errors
- Message types: UserMessage, AssistantMessage, SystemMessage, ResultMessage, StreamEvent
- Content blocks: TextBlock, ThinkingBlock, ToolUseBlock, ToolResultBlock
- Tool builder API for creating custom tools
- Calculator toolkit example
- Extensive examples covering all features
- Complete test coverage
- Documentation and architecture guides

### Features
- Streaming and non-streaming modes
- Permission modes: default, acceptEdits, plan, bypassPermissions
- Control protocol for bidirectional communication
- MCP message routing
- Hook callbacks with regex matchers
- Budget limits and usage tracking
- Beta feature support
- Custom system prompts
- Environment variable configuration
- Verbose logging

[Unreleased]: https://github.com/M1n9X/claude-agent-sdk-go/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/M1n9X/claude-agent-sdk-go/releases/tag/v0.1.0
