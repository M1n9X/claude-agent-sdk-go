package types

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// McpTool represents a tool that can be executed by Claude.
// Tools are registered with MCP servers and can be invoked by Claude
// during conversations.
type McpTool interface {
	// Name returns the tool name (unique within a server).
	Name() string

	// Description returns a human-readable description of what the tool does.
	Description() string

	// InputSchema returns the JSON schema for tool parameters.
	InputSchema() map[string]interface{}

	// Execute runs the tool with the given input.
	// Returns a ToolResult containing the execution result.
	// If execution fails, returns an error.
	Execute(ctx context.Context, input map[string]interface{}) (*ToolResult, error)
}

// ToolResult represents the result of tool execution.
type ToolResult struct {
	// Content is the result content (array of content blocks).
	Content []ContentBlock `json:"content"`

	// IsError indicates if the result represents an error.
	IsError bool `json:"isError,omitempty"`
}


// ToolFunc is the function signature for tool handler functions.
type ToolFunc func(
	ctx context.Context,
	input map[string]interface{},
) (*ToolResult, error)

// ToolBuilder builds a tool using the builder pattern.
// Provides a fluent API for defining tools with parameters,
// validation, and handlers.
type ToolBuilder struct {
	name        string
	description string
	params      []ToolParam
	required    []string
	handler     ToolFunc
	validator   func(map[string]interface{}) error
	enums       map[string][]interface{}
}

// ToolParam represents a parameter definition for a tool.
type ToolParam struct {
	Name        string
	Type        string
	Description string
	Required    bool
	Enum        []interface{}
	Default     interface{}
	Properties  map[string]ToolParam // for object types
}

// NewTool creates a new tool builder.
// The tool name should be unique within an MCP server.
func NewTool(name string) *ToolBuilder {
	return &ToolBuilder{
		name:   name,
		params: []ToolParam{},
		required: []string{},
		enums: make(map[string][]interface{}),
	}
}

// Description sets the tool description.
func (b *ToolBuilder) Description(desc string) *ToolBuilder {
	b.description = desc
	return b
}

// StringParam adds a string parameter.
func (b *ToolBuilder) StringParam(name, desc string, required bool) *ToolBuilder {
	b.addParam(ToolParam{
		Name:        name,
		Type:        "string",
		Description: desc,
		Required:    required,
	}, required)
	return b
}

// NumberParam adds a number (float) parameter.
func (b *ToolBuilder) NumberParam(name, desc string, required bool) *ToolBuilder {
	b.addParam(ToolParam{
		Name:        name,
		Type:        "number",
		Description: desc,
		Required:    required,
	}, required)
	return b
}

// IntParam adds an integer parameter.
func (b *ToolBuilder) IntParam(name, desc string, required bool) *ToolBuilder {
	b.addParam(ToolParam{
		Name:        name,
		Type:        "integer",
		Description: desc,
		Required:    required,
	}, required)
	return b
}

// BoolParam adds a boolean parameter.
func (b *ToolBuilder) BoolParam(name, desc string, required bool) *ToolBuilder {
	b.addParam(ToolParam{
		Name:        name,
		Type:        "boolean",
		Description: desc,
		Required:    required,
	}, required)
	return b
}

// ArrayParam adds an array parameter.
func (b *ToolBuilder) ArrayParam(name, desc string, required bool, itemType string) *ToolBuilder {
	b.addParam(ToolParam{
		Name:        name,
		Type:        "array",
		Description: desc,
		Required:    required,
		Default:     map[string]string{"itemsType": itemType},
	}, required)
	return b
}

// EnumParam adds an enum parameter (string with allowed values).
func (b *ToolBuilder) EnumParam(name, desc string, required bool, enum []interface{}) *ToolBuilder {
	b.enums[name] = enum
	b.addParam(ToolParam{
		Name:        name,
		Type:        "string",
		Description: desc,
		Required:    required,
		Enum:        enum,
	}, required)
	return b
}

// ObjectParam adds an object parameter with nested properties.
func (b *ToolBuilder) ObjectParam(name, desc string, required bool, properties map[string]ToolParam) *ToolBuilder {
	b.addParam(ToolParam{
		Name:        name,
		Type:        "object",
		Description: desc,
		Required:    required,
		Properties:  properties,
	}, required)
	return b
}

// ObjectArrayParam adds an array parameter containing objects.
// This is useful for lists of complex items where each item has multiple properties.
func (b *ToolBuilder) ObjectArrayParam(name, desc string, required bool, itemSchema map[string]ToolParam) *ToolBuilder {
	b.addParam(ToolParam{
		Name:        name,
		Type:        "array",
		Description: desc,
		Required:    required,
		Default: map[string]interface{}{
			"items": map[string]interface{}{
				"type":       "object",
				"properties": itemSchema,
			},
		},
	}, required)
	return b
}

// DefaultParam sets a default value for the last added parameter.
// The parameter must already be added.
func (b *ToolBuilder) DefaultParam(name string, defaultValue interface{}) *ToolBuilder {
	for i := range b.params {
		if b.params[i].Name == name {
			b.params[i].Default = defaultValue
			break
		}
	}
	return b
}

// addParam adds a parameter to the list and tracks required status.
// This is an internal helper method.
func (b *ToolBuilder) addParam(param ToolParam, required bool) {
	b.params = append(b.params, param)
	if required {
		b.required = append(b.required, param.Name)
	}
}

// Handler sets the tool handler function.
func (b *ToolBuilder) Handler(fn ToolFunc) *ToolBuilder {
	b.handler = fn
	return b
}

// WithValidation adds a custom validation function.
func (b *ToolBuilder) WithValidation(fn func(map[string]interface{}) error) *ToolBuilder {
	b.validator = fn
	return b
}

// Build constructs the tool.
// Returns an error if required fields are missing or validation fails.
func (b *ToolBuilder) Build() (McpTool, error) {
	if b.name == "" {
		return nil, fmt.Errorf("tool name is required")
	}
	if b.description == "" {
		return nil, fmt.Errorf("tool description is required")
	}
	if b.handler == nil {
		return nil, fmt.Errorf("tool handler is required")
	}

	schema := b.buildJSONSchema()

	return &tool{
		name:        b.name,
		description: b.description,
		inputSchema: schema,
		handler:     b.handler,
		validator:   b.validator,
	}, nil
}

// buildJSONSchema constructs the JSON schema from parameters.
func (b *ToolBuilder) buildJSONSchema() map[string]interface{} {
	schema := map[string]interface{}{
		"type":       "object",
		"properties": make(map[string]interface{}),
		"required":   b.required,
	}

	properties := schema["properties"].(map[string]interface{})

	for _, param := range b.params {
		prop := map[string]interface{}{
			"type":        param.Type,
			"description": param.Description,
		}

		// Add enum if present
		if len(param.Enum) > 0 {
			prop["enum"] = param.Enum
		}

		// Add default if present
		if param.Default != nil {
			prop["default"] = param.Default
		}

		// Add items for arrays
		if param.Type == "array" {
			if itemType, ok := param.Default.(map[string]string); ok {
				prop["items"] = map[string]string{
					"type": itemType["itemsType"],
				}
			}
		}

		// Add properties for objects
		if param.Type == "object" && param.Properties != nil {
			objProps := make(map[string]interface{})
			for k, v := range param.Properties {
				objProps[k] = map[string]interface{}{
					"type":        v.Type,
					"description": v.Description,
				}
			}
			prop["properties"] = objProps

			// Handle required fields in nested object
			var nestedRequired []string
			for _, prop := range param.Properties {
				if prop.Required {
					nestedRequired = append(nestedRequired, prop.Name)
				}
			}
			if len(nestedRequired) > 0 {
				prop["required"] = nestedRequired
			}
		}

		properties[param.Name] = prop
	}

	return schema
}

// tool implements the McpTool interface.
type tool struct {
	name        string
	description string
	inputSchema map[string]interface{}
	handler     ToolFunc
	validator   func(map[string]interface{}) error
}

func (t *tool) Name() string {
	return t.name
}

func (t *tool) Description() string {
	return t.description
}

func (t *tool) InputSchema() map[string]interface{} {
	return t.inputSchema
}

func (t *tool) Execute(ctx context.Context, input map[string]interface{}) (*ToolResult, error) {
	// Validate input against schema
	if err := validateInput(t.inputSchema, input); err != nil {
		return nil, fmt.Errorf("input validation failed: %w", err)
	}

	// Run custom validator if present
	if t.validator != nil {
		if err := t.validator(input); err != nil {
			return nil, fmt.Errorf("custom validation failed: %w", err)
		}
	}

	return t.handler(ctx, input)
}

// validateInput validates input against JSON schema.
func validateInput(schema map[string]interface{}, input map[string]interface{}) error {
	// Validate type
	schemaType, ok := schema["type"].(string)
	if !ok || schemaType != "object" {
		return fmt.Errorf("invalid schema type: %v", schemaType)
	}

	// Validate required fields
	if required, ok := schema["required"].([]string); ok {
		for _, field := range required {
			if _, exists := input[field]; !exists {
				return fmt.Errorf("missing required field: %s", field)
			}
		}
	}

	// Validate properties
	properties, ok := schema["properties"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid schema properties")
	}

	for key, value := range input {
		propSchema, exists := properties[key]
		if !exists {
			return fmt.Errorf("unknown field: %s", key)
		}

		prop, ok := propSchema.(map[string]interface{})
		if !ok {
			continue // Skip validation if property schema is malformed
		}

		propType, ok := prop["type"].(string)
		if !ok {
			continue
		}

		// Type validation
		switch propType {
		case "string":
			if _, ok := value.(string); !ok {
				return fmt.Errorf("field %s must be string, got %T", key, value)
			}
		case "number":
			if _, ok := value.(float64); !ok {
				return fmt.Errorf("field %s must be number, got %T", key, value)
			}
		case "integer":
			// JSON unmarshals integers as float64
			if f, ok := value.(float64); !ok || f != float64(int64(f)) {
				return fmt.Errorf("field %s must be integer, got %T", key, value)
			}
		case "boolean":
			if _, ok := value.(bool); !ok {
				return fmt.Errorf("field %s must be boolean, got %T", key, value)
			}
		case "array":
			if _, ok := value.([]interface{}); !ok {
				return fmt.Errorf("field %s must be array, got %T", key, value)
			}
		case "object":
			if objValue, ok := value.(map[string]interface{}); ok {
				// Recursively validate nested object
				if nestedProps, ok := prop["properties"].(map[string]interface{}); ok {
					nestedSchema := map[string]interface{}{
						"type":       "object",
						"properties": nestedProps,
					}
					if required, ok := prop["required"].([]string); ok {
						nestedSchema["required"] = required
					}
					if err := validateInput(nestedSchema, objValue); err != nil {
						return fmt.Errorf("nested validation failed for %s: %w", key, err)
					}
				}
			} else {
				return fmt.Errorf("field %s must be object, got %T", key, value)
			}
		}

		// Enum validation
		if enum, ok := prop["enum"].([]interface{}); ok {
			valid := false
			for _, e := range enum {
				if value == e {
					valid = true
					break
				}
			}
			if !valid {
				enumJSON, _ := json.Marshal(enum)
				return fmt.Errorf("field %s must be one of %s, got %v", key, enumJSON, value)
			}
		}
	}

	return nil
}

// NewMcpToolResult creates a successful tool result.
func NewMcpToolResult(content ...ContentBlock) *ToolResult {
	return &ToolResult{
		Content: content,
		IsError: false,
	}
}

// NewErrorMcpToolResult creates an error tool result.
func NewErrorMcpToolResult(message string) *ToolResult {
	return &ToolResult{
		Content: []ContentBlock{
			TextBlock{
				Type: "text",
				Text: message,
			},
		},
		IsError: true,
	}
}

// Helper functions for common tool creation

// NewFileReadTool creates a file reading tool.
// Reads content from a file at the specified path.
func NewFileReadTool() (McpTool, error) {
	return NewTool("read_file").
		Description("Read content from a file").
		StringParam("path", "Path to the file", true).
		Handler(func(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
			path := args["path"].(string)

			// Check if path is safe (simplified)
			if strings.Contains(path, "..") {
				return NewErrorMcpToolResult("Invalid file path: path traversal detected"), nil
			}

			content, err := os.ReadFile(path)
			if err != nil {
				return NewErrorMcpToolResult(
					fmt.Sprintf("Failed to read file %s: %v", path, err),
				), nil
			}

			return NewMcpToolResult(
				TextBlock{
					Type: "text",
					Text: string(content),
				},
			), nil
		}).
		Build()
}

// NewFileWriteTool creates a file writing tool.
// Writes content to a file at the specified path.
func NewFileWriteTool() (McpTool, error) {
	return NewTool("write_file").
		Description("Write content to a file").
		StringParam("path", "Path to the file", true).
		StringParam("content", "Content to write", true).
		Handler(func(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
			path := args["path"].(string)
			content := args["content"].(string)

			// Check if path is safe
			if strings.Contains(path, "..") {
				return NewErrorMcpToolResult("Invalid file path: path traversal detected"), nil
			}

			// Ensure directory exists
			dir := filepath.Dir(path)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return NewErrorMcpToolResult(
					fmt.Sprintf("Failed to create directory %s: %v", dir, err),
				), nil
			}

			err := os.WriteFile(path, []byte(content), 0644)
			if err != nil {
				return NewErrorMcpToolResult(
					fmt.Sprintf("Failed to write file %s: %v", path, err),
				), nil
			}

			return NewMcpToolResult(
				TextBlock{
					Type: "text",
					Text: fmt.Sprintf("Successfully wrote %d bytes to %s", len(content), path),
				},
			), nil
		}).
		Build()
}

// Calculator Example - Complete toolkit

// NewCalculatorToolkit creates a suite of calculator tools.
func NewCalculatorToolkit() ([]McpTool, error) {
	var tools []McpTool

	// Add tool
	addTool, err := NewTool("add").
		Description("Add two numbers").
		NumberParam("a", "First number", true).
		NumberParam("b", "Second number", true).
		Handler(func(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
			a := args["a"].(float64)
			b := args["b"].(float64)
			result := a + b

			return NewMcpToolResult(
				TextBlock{
					Type: "text",
					Text: fmt.Sprintf("%.2f + %.2f = %.2f", a, b, result),
				},
			), nil
		}).
		Build()
	if err != nil {
		return nil, err
	}
	tools = append(tools, addTool)

	// Subtract tool
	subTool, err := NewTool("subtract").
		Description("Subtract second number from first").
		NumberParam("a", "First number", true).
		NumberParam("b", "Second number", true).
		Handler(func(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
			a := args["a"].(float64)
			b := args["b"].(float64)
			result := a - b

			return NewMcpToolResult(
				TextBlock{
					Type: "text",
					Text: fmt.Sprintf("%.2f - %.2f = %.2f", a, b, result),
				},
			), nil
		}).
		Build()
	if err != nil {
		return nil, err
	}
	tools = append(tools, subTool)

	// Multiply tool
	mulTool, err := NewTool("multiply").
		Description("Multiply two numbers").
		NumberParam("a", "First number", true).
		NumberParam("b", "Second number", true).
		Handler(func(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
			a := args["a"].(float64)
			b := args["b"].(float64)
			result := a * b

			return NewMcpToolResult(
				TextBlock{
					Type: "text",
					Text: fmt.Sprintf("%.2f * %.2f = %.2f", a, b, result),
				},
			), nil
		}).
		Build()
	if err != nil {
		return nil, err
	}
	tools = append(tools, mulTool)

	// Divide tool with error handling
	divTool, err := NewTool("divide").
		Description("Divide first number by second").
		NumberParam("a", "Dividend", true).
		NumberParam("b", "Divisor", true).
		Handler(func(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
			a := args["a"].(float64)
			b := args["b"].(float64)

			if b == 0 {
				return NewErrorMcpToolResult("Division by zero is not allowed"), nil
			}

			result := a / b

			return NewMcpToolResult(
				TextBlock{
					Type: "text",
					Text: fmt.Sprintf("%.2f / %.2f = %.2f", a, b, result),
				},
			), nil
		}).
		Build()
	if err != nil {
		return nil, err
	}
	tools = append(tools, divTool)

	// Power tool
	powTool, err := NewTool("power").
		Description("Raise a number to a power").
		NumberParam("base", "Base number", true).
		NumberParam("exponent", "Exponent", true).
		Handler(func(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
			base := args["base"].(float64)
			exponent := args["exponent"].(float64)
			result := 1.0

			// Simple power implementation (handles integer exponents well)
			if exponent >= 0 {
				for i := 0; i < int(exponent); i++ {
					result *= base
				}
			} else {
				for i := 0; i < int(-exponent); i++ {
					result /= base
				}
			}

			return NewMcpToolResult(
				TextBlock{
					Type: "text",
					Text: fmt.Sprintf("%.2f ^ %.2f = %.2f", base, exponent, result),
				},
			), nil
		}).
		Build()
	if err != nil {
		return nil, err
	}
	tools = append(tools, powTool)

	return tools, nil
}

// ToolManager manages a collection of tools and can create MCP servers.
type ToolManager struct {
	tools map[string]McpTool
	mu    sync.RWMutex
}

// NewToolManager creates a new tool manager.
func NewToolManager() *ToolManager {
	return &ToolManager{
		tools: make(map[string]McpTool),
	}
}

// Register registers a tool with the manager.
// Returns an error if a tool with the same name already exists.
func (m *ToolManager) Register(tool McpTool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.tools[tool.Name()]; exists {
		return fmt.Errorf("tool already registered: %s", tool.Name())
	}

	m.tools[tool.Name()] = tool
	return nil
}

// MustRegister registers a tool and panics if it fails.
// Useful for initialization code.
func (m *ToolManager) MustRegister(tool McpTool) {
	if err := m.Register(tool); err != nil {
		panic(err)
	}
}

// Get retrieves a tool by name.
func (m *ToolManager) Get(name string) (McpTool, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tool, exists := m.tools[name]
	return tool, exists
}

// List returns all registered tools.
func (m *ToolManager) List() []McpTool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tools := make([]McpTool, 0, len(m.tools))
	for _, tool := range m.tools {
		tools = append(tools, tool)
	}
	return tools
}

// Names returns the names of all registered tools.
func (m *ToolManager) Names() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.tools))
	for name := range m.tools {
		names = append(names, name)
	}
	return names
}

// Count returns the number of registered tools.
func (m *ToolManager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.tools)
}

// Clear removes all registered tools.
func (m *ToolManager) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tools = make(map[string]McpTool)
}

// Unregister removes a tool from the manager.
func (m *ToolManager) Unregister(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.tools[name]; !exists {
		return fmt.Errorf("tool not found: %s", name)
	}

	delete(m.tools, name)
	return nil
}

// CreateServer creates an MCP server with all registered tools.
// This is a convenience method for creating servers from the tool manager.
func (m *ToolManager) CreateServer(name, version string) *ToolServerConfig {
	tools := m.List()
	return CreateToolServer(name, version, tools)
}

// ToolServerConfig is the configuration for an SDK MCP server.
type ToolServerConfig struct {
	Type     string      `json:"type"`
	Name     string      `json:"name"`
	Version  string      `json:"version,omitempty"`
	Instance interface{} `json:"instance"`
}

// CreateToolServer creates an SDK MCP server configuration.
// This function is the equivalent of Python's create_sdk_mcp_server().
func CreateToolServer(name, version string, tools []McpTool) *ToolServerConfig {
	return &ToolServerConfig{
		Type:     "sdk",
		Name:     name,
		Version:  version,
		Instance: tools, // Actual server implementation will be in internal/mcp
	}
}
