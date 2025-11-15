package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/M1n9X/claude-agent-sdk-go/internal/mcp"
	"github.com/M1n9X/claude-agent-sdk-go/types"
)

// MCPServerTransport wraps a transport with MCP server support.
// It handles routing between SDK MCP servers (in-process) and external transports.
type MCPServerTransport struct {
	// The underlying transport (subprocess, HTTP, etc.)
	transport Transport

	// SDK MCP servers registered in-process
	sdkMCPServers map[string]*mcp.SdkMCPServer
	sdkMCPMu      sync.RWMutex

	// MCP configuration file path (will be cleaned up on Close)
	mcpConfigFile string

	// CanUseTool callback for permission control
	canUseTool types.CanUseToolFunc

	ctx    context.Context
	cancel context.CancelFunc
}

// NewMCPServerTransport creates a new MCP server transport wrapper.
func NewMCPServerTransport(transport Transport, options *types.ClaudeAgentOptions) *MCPServerTransport {
	ctx, cancel := context.WithCancel(context.Background())

	t := &MCPServerTransport{
		transport:     transport,
		sdkMCPServers: make(map[string]*mcp.SdkMCPServer),
		ctx:           ctx,
		cancel:        cancel,
		canUseTool:    options.CanUseTool,
	}

	// Initialize SDK MCP servers from options
	t.initializeMCPServers(options)

	return t
}

// initializeMCPServers initializes SDK MCP servers from options.
func (t *MCPServerTransport) initializeMCPServers(options *types.ClaudeAgentOptions) {
	if options.McpServers == nil {
		return
	}

	// Type assert to the expected map type
	if servers, ok := options.McpServers.(map[string]interface{}); ok {
		for name, config := range servers {
			// Check if it's an SDK MCP server
			if _, ok := config.(*types.ToolServerConfig); ok {
				// The actual server instance will be created later
				t.sdkMCPServers[name] = nil // Placeholder
			}
		}
	}
}

// Connect establishes connection and initializes MCP servers.
func (t *MCPServerTransport) Connect(ctx context.Context) error {
	if err := t.transport.Connect(ctx); err != nil {
		return fmt.Errorf("transport connect: %w", err)
	}
	return nil
}

// Close closes the transport and cleans up MCP resources.
func (t *MCPServerTransport) Close(ctx context.Context) error {
	t.cancel()

	// Cleanup MCP config file if exists
	if t.mcpConfigFile != "" {
		os.Remove(t.mcpConfigFile)
		t.mcpConfigFile = ""
	}

	return t.transport.Close(ctx)
}

// Write writes a message to the transport.
func (t *MCPServerTransport) Write(ctx context.Context, data string) error {
	return t.transport.Write(ctx, data)
}

// ReadMessages returns a channel of incoming messages.
func (t *MCPServerTransport) ReadMessages(ctx context.Context) <-chan types.Message {
	return t.transport.ReadMessages(ctx)
}

// OnError handles errors from the transport.
func (t *MCPServerTransport) OnError(err error) {
	t.transport.OnError(err)
}

// IsReady checks if the transport is ready.
func (t *MCPServerTransport) IsReady() bool {
	return t.transport.IsReady()
}

// GetError returns any error from the transport.
func (t *MCPServerTransport) GetError() error {
	return t.transport.GetError()
}

// RegisterSDKMCPServer registers an SDK MCP server with the transport.
func (t *MCPServerTransport) RegisterSDKMCPServer(name string, server *mcp.SdkMCPServer) {
	t.sdkMCPMu.Lock()
	defer t.sdkMCPMu.Unlock()

	t.sdkMCPServers[name] = server
}

// GetSDKMCPServer retrieves an SDK MCP server by name.
func (t *MCPServerTransport) GetSDKMCPServer(name string) (*mcp.SdkMCPServer, bool) {
	t.sdkMCPMu.RLock()
	defer t.sdkMCPMu.RUnlock()

	server, exists := t.sdkMCPServers[name]
	return server, exists
}

// RouteToolUse routes a tool use request to the appropriate server.
// Returns: isMcp, serverName, toolName, error
func (t *MCPServerTransport) RouteToolUse(toolName string) (bool, string, string, error) {
	// Tool name format: "mcp__server__tool" or "tool"
	parts := splitToolName(toolName)

	if len(parts) == 3 && parts[0] == "mcp" {
		// MCP tool: mcp__server__tool
		serverName := parts[1]
		toolName := parts[2]

		// Check if it's a registered SDK MCP server
		if _, exists := t.GetSDKMCPServer(serverName); exists {
			return true, serverName, toolName, nil
		}

		// Otherwise, it's an external MCP server
		return true, serverName, toolName, nil
	}

	// Regular CLI tool
	return false, "", toolName, nil
}

// splitToolName splits a tool name into parts.
func splitToolName(toolName string) []string {
	// Tool name format: "mcp__server__tool" or "tool"
	// Split by "__" (double underscore)
	var parts []string
	current := ""

	for i := 0; i < len(toolName); i++ {
		if i < len(toolName)-1 && toolName[i] == '_' && toolName[i+1] == '_' {
			// Found double underscore
			parts = append(parts, current)
			current = ""
			i++ // Skip second underscore
		} else {
			current += string(toolName[i])
		}
	}

	if current != "" {
		parts = append(parts, current)
	}

	return parts
}

// generateMcpConfigFile generates a temporary MCP configuration file.
func (t *MCPServerTransport) generateMcpConfigFile(options *types.ClaudeAgentOptions) (string, error) {
	if options.McpServers == nil {
		return "", nil // No MCP servers configured
	}

	// Type assert to the expected map type
	servers, ok := options.McpServers.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid McpServers type, expected map[string]interface{}")
	}

	if len(servers) == 0 {
		return "", nil
	}

	config := map[string]interface{}{
		"mcpServers": make(map[string]interface{}),
	}

	for name, server := range servers {
		if toolConfig, ok := server.(*types.ToolServerConfig); ok && toolConfig.Type == "sdk" {
			// SDK MCP server - handled in-process, not in config file
			continue
		}

		// Add to config based on type
		switch s := server.(type) {
		case types.McpStdioServerConfig:
			config["mcpServers"].(map[string]interface{})[name] = map[string]interface{}{
				"type":    "stdio",
				"command": s.Command,
				"args":    s.Args,
				"env":     s.Env,
			}
		case types.McpSSEServerConfig:
			config["mcpServers"].(map[string]interface{})[name] = map[string]interface{}{
				"type":    "sse",
				"url":     s.URL,
				"headers": s.Headers,
			}
		case types.McpHTTPServerConfig:
			config["mcpServers"].(map[string]interface{})[name] = map[string]interface{}{
				"type":    "http",
				"url":     s.URL,
				"headers": s.Headers,
			}
		}
	}

	// If no external servers, return empty string
	if len(config["mcpServers"].(map[string]interface{})) == 0 {
		return "", nil
	}

	// Create temp file
	tmpFile, err := os.CreateTemp("", "claude-mcp-*.json")
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}
	defer tmpFile.Close()

	encoder := json.NewEncoder(tmpFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(config); err != nil {
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("encode config: %w", err)
	}

	return tmpFile.Name(), nil
}

// BuildMcpCommand augments the CLI command with MCP configuration.
func BuildMcpCommand(baseCmd []string, options *types.ClaudeAgentOptions) ([]string, error) {
	cmd := append([]string{}, baseCmd...)

	// Generate MCP config file
	transport := &MCPServerTransport{
		sdkMCPServers: make(map[string]*mcp.SdkMCPServer),
	}

	configFile, err := transport.generateMcpConfigFile(options)
	if err != nil {
		return nil, fmt.Errorf("generate MCP config: %w", err)
	}

	if configFile != "" {
		cmd = append(cmd, "--mcp-servers", configFile)
	}

	return cmd, nil
}
