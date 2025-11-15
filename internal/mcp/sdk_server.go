// Package mcp implements the Model Context Protocol (MCP) for Claude Code.
// This package provides SDK MCP server functionality for in-process tool execution.
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/M1n9X/claude-agent-sdk-go/types"
)

// SdkMCPServer implements an in-process MCP server for executing tools.
// It handles MCP protocol messages and routes tool calls to registered tools.
type SdkMCPServer struct {
	name     string
	version  string
	tools    []types.McpTool
	toolsMap map[string]types.McpTool // name -> tool for fast lookup
	mu       sync.RWMutex             // protects tools and toolsMap
}

// NewSdkMCPServer creates a new SDK MCP server instance.
// The name and version identify the server, and tools are the initial set of tools.
func NewSdkMCPServer(name, version string, tools []types.McpTool) *SdkMCPServer {
	server := &SdkMCPServer{
		name:     name,
		version:  version,
		tools:    tools,
		toolsMap: make(map[string]types.McpTool),
	}

	// Index tools by name for fast lookup
	for _, tool := range tools {
		server.toolsMap[tool.Name()] = tool
	}

	return server
}

// Name returns the server name.
func (s *SdkMCPServer) Name() string {
	return s.name
}

// Version returns the server version.
func (s *SdkMCPServer) Version() string {
	return s.version
}

// Tools returns all registered tools.
func (s *SdkMCPServer) Tools() []types.McpTool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.tools
}

// AddTool adds a new tool to the server.
// Returns an error if a tool with the same name already exists.
func (s *SdkMCPServer) AddTool(tool types.McpTool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.toolsMap[tool.Name()]; exists {
		return fmt.Errorf("tool already exists: %s", tool.Name())
	}

	s.tools = append(s.tools, tool)
	s.toolsMap[tool.Name()] = tool

	return nil
}

// RemoveTool removes a tool from the server.
// Returns an error if the tool doesn't exist.
func (s *SdkMCPServer) RemoveTool(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.toolsMap[name]; !exists {
		return fmt.Errorf("tool not found: %s", name)
	}

	// Remove from map
	delete(s.toolsMap, name)

	// Remove from slice
	for i, t := range s.tools {
		if t.Name() == name {
			s.tools = append(s.tools[:i], s.tools[i+1:]...)
			break
		}
	}

	return nil
}

// HandleMessage processes an MCP JSON-RPC message and returns a response.
// This is the main entry point for handling MCP protocol messages.
func (s *SdkMCPServer) HandleMessage(msg map[string]interface{}) (map[string]interface{}, error) {
	method, ok := msg["method"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid method field")
	}

	switch method {
	case "initialize":
		return s.handleInitialize(msg)
	case "tools/list":
		return s.handleToolsList(msg)
	case "tools/call":
		return s.handleToolsCall(msg)
	default:
		id := msg["id"]
		resp := NewErrorResponse(id, ErrorCodeMethodNotFound, fmt.Sprintf("method not found: %s", method))
		return responseToMap(resp), nil
	}
}

// handleInitialize handles the initialize request from the client.
func (s *SdkMCPServer) handleInitialize(msg map[string]interface{}) (map[string]interface{}, error) {
	id := msg["id"]

	result := map[string]interface{}{
		"protocolVersion": "0.1.0",
		"capabilities": map[string]interface{}{
			"tools": map[string]interface{}{
				"listChanged": false,
			},
		},
		"serverInfo": map[string]interface{}{
			"name":    s.name,
			"version": s.version,
		},
	}

	resp := NewSuccessResponse(id, result)
	return responseToMap(resp), nil
}

// handleToolsList handles the tools/list request.
func (s *SdkMCPServer) handleToolsList(msg map[string]interface{}) (map[string]interface{}, error) {
	id := msg["id"]

	tools := s.Tools()
	toolList := make([]map[string]interface{}, len(tools))

	for i, tool := range tools {
		toolList[i] = map[string]interface{}{
			"name":        tool.Name(),
			"description": tool.Description(),
			"inputSchema": tool.InputSchema(),
		}
	}

	result := map[string]interface{}{
		"tools": toolList,
	}

	resp := NewSuccessResponse(id, result)
	return responseToMap(resp), nil
}

// handleToolsCall handles a tools/call request.
func (s *SdkMCPServer) handleToolsCall(msg map[string]interface{}) (map[string]interface{}, error) {
	id := msg["id"]

	params, ok := msg["params"].(map[string]interface{})
	if !ok {
		errResp := NewErrorResponse(id, ErrorCodeInvalidParams, "missing or invalid params")
		return responseToMap(errResp), nil
	}

	name, ok := params["name"].(string)
	if !ok {
		errResp := NewErrorResponse(id, ErrorCodeInvalidParams, "missing or invalid tool name")
		return responseToMap(errResp), nil
	}

	tool, exists := s.toolsMap[name]
	if !exists {
		errResp := NewErrorResponse(id, ErrorCodeMethodNotFound, fmt.Sprintf("tool not found: %s", name))
		return responseToMap(errResp), nil
	}

	input, ok := params["arguments"].(map[string]interface{})
	if !ok {
		errResp := NewErrorResponse(id, ErrorCodeInvalidParams, "missing or invalid arguments")
		return responseToMap(errResp), nil
	}

	// Execute the tool
	ctx := context.Background()
	result, err := tool.Execute(ctx, input)
	if err != nil {
		errResp := NewErrorResponse(id, ErrorCodeInternalError, fmt.Sprintf("tool execution failed: %v", err))
		return responseToMap(errResp), nil
	}

	resp := NewSuccessResponse(id, result)
	return responseToMap(resp), nil
}

// responseToMap converts a Response to a map for JSON serialization.
func responseToMap(resp *Response) map[string]interface{} {
	result := map[string]interface{}{
		"jsonrpc": resp.JsonRpc,
		"id":      resp.ID,
	}

	if resp.Error != nil {
		errMap := map[string]interface{}{
			"code":    resp.Error.Code,
			"message": resp.Error.Message,
		}
		if resp.Error.Data != nil {
			errMap["data"] = resp.Error.Data
		}
		result["error"] = errMap
	} else {
		// Marshal and unmarshal the result to ensure proper JSON format
		// This ensures that complex types like ToolResult are properly serialized
		// as maps/arrays rather than Go structs for MCP protocol compatibility
		if resp.Result != nil {
			data, err := json.Marshal(resp.Result)
			if err != nil {
				// If marshaling fails, return the original result as fallback
				result["result"] = resp.Result
			} else {
				var unmarshaledResult interface{}
				if err := json.Unmarshal(data, &unmarshaledResult); err != nil {
					// If unmarshaling fails, return the original result as fallback
					result["result"] = resp.Result
				} else {
					result["result"] = unmarshaledResult
				}
			}
		}
	}

	return result
}

// ToolServerConfig represents SDK MCP server configuration.
// This is used to configure an in-process MCP server in ClaudeAgentOptions.
type ToolServerConfig struct {
	Type     string      `json:"type"` // Always "sdk"
	Name     string      `json:"name"`
	Version  string      `json:"version,omitempty"`
	Instance interface{} `json:"instance,omitempty"` // *SdkMCPServer
}

// CreateSdkMCPServer creates a ToolServerConfig for an SDK MCP server.
// This makes it easy to register SDK MCP servers in ClaudeAgentOptions.
func CreateSdkMCPServer(name, version string, tools []types.McpTool) *ToolServerConfig {
	return &ToolServerConfig{
		Type:     "sdk",
		Name:     name,
		Version:  version,
		Instance: NewSdkMCPServer(name, version, tools),
	}
}
