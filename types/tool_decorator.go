package types

import (
	"fmt"
)

// SimpleTool provides a decorator-style API for defining tools,
// similar to Python's @tool decorator.
//
// Example usage:
//
//	greetTool := types.SimpleTool{
//	    Name:        "greet",
//	    Description: "Greet a user by name",
//	    Parameters: map[string]types.SimpleParam{
//	        "name": {Type: "string", Description: "User's name", Required: true},
//	    },
//	    Handler: func(ctx context.Context, args map[string]interface{}) (*types.ToolResult, error) {
//	        name := args["name"].(string)
//	        return types.NewMcpToolResult(
//	            types.TextBlock{
//	                Type: "text",
//	                Text: fmt.Sprintf("Hello, %s!", name),
//	            },
//	        ), nil
//	    },
//	}
//
//	tool, err := greetTool.Build()
type SimpleTool struct {
	Name        string
	Description string
	Parameters  map[string]SimpleParam
	Handler     ToolFunc
}

// SimpleParam represents a simplified parameter definition.
type SimpleParam struct {
	Type        string // "string", "number", "integer", "boolean", "array", "object"
	Description string
	Required    bool
	Enum        []interface{}          // For enum types
	Default     interface{}            // Default value
	Items       *SimpleParam           // For array types
	Properties  map[string]SimpleParam // For object types
}

// Build converts a SimpleTool into an McpTool.
func (s *SimpleTool) Build() (McpTool, error) {
	if s.Name == "" {
		return nil, fmt.Errorf("tool name is required")
	}
	if s.Description == "" {
		return nil, fmt.Errorf("tool description is required")
	}
	if s.Handler == nil {
		return nil, fmt.Errorf("tool handler is required")
	}

	// Build JSON schema from parameters
	schema := map[string]interface{}{
		"type":       "object",
		"properties": make(map[string]interface{}),
	}

	properties := schema["properties"].(map[string]interface{})
	var required []string

	for name, param := range s.Parameters {
		prop := buildParamSchema(param)
		properties[name] = prop

		if param.Required {
			required = append(required, name)
		}
	}

	if len(required) > 0 {
		schema["required"] = required
	}

	return &tool{
		name:        s.Name,
		description: s.Description,
		inputSchema: schema,
		handler:     s.Handler,
	}, nil
}

// buildParamSchema builds a JSON schema for a parameter.
func buildParamSchema(param SimpleParam) map[string]interface{} {
	prop := map[string]interface{}{
		"type":        param.Type,
		"description": param.Description,
	}

	if len(param.Enum) > 0 {
		prop["enum"] = param.Enum
	}

	if param.Default != nil {
		prop["default"] = param.Default
	}

	// Handle array items
	if param.Type == "array" && param.Items != nil {
		prop["items"] = buildParamSchema(*param.Items)
	}

	// Handle object properties
	if param.Type == "object" && len(param.Properties) > 0 {
		objProps := make(map[string]interface{})
		var objRequired []string

		for propName, propParam := range param.Properties {
			objProps[propName] = buildParamSchema(propParam)
			if propParam.Required {
				objRequired = append(objRequired, propName)
			}
		}

		prop["properties"] = objProps
		if len(objRequired) > 0 {
			prop["required"] = objRequired
		}
	}

	return prop
}

// ToolDecorator provides a fluent API for creating tools with a decorator-like syntax.
// This is the Go equivalent of Python's @tool decorator.
//
// Example usage:
//
//	tool := types.Tool("greet", "Greet a user").
//	    Param("name", "string", "User's name", true).
//	    Handle(func(ctx context.Context, args map[string]interface{}) (*types.ToolResult, error) {
//	        name := args["name"].(string)
//	        return types.NewMcpToolResult(
//	            types.TextBlock{Type: "text", Text: fmt.Sprintf("Hello, %s!", name)},
//	        ), nil
//	    })
type ToolDecorator struct {
	name        string
	description string
	params      map[string]SimpleParam
	handler     ToolFunc
}

// Tool creates a new tool decorator with the given name and description.
// This is the entry point for the decorator-style API.
func Tool(name, description string) *ToolDecorator {
	return &ToolDecorator{
		name:        name,
		description: description,
		params:      make(map[string]SimpleParam),
	}
}

// Param adds a parameter to the tool.
func (d *ToolDecorator) Param(name, paramType, description string, required bool) *ToolDecorator {
	d.params[name] = SimpleParam{
		Type:        paramType,
		Description: description,
		Required:    required,
	}
	return d
}

// EnumParam adds an enum parameter to the tool.
func (d *ToolDecorator) EnumParam(name, description string, required bool, enum []interface{}) *ToolDecorator {
	d.params[name] = SimpleParam{
		Type:        "string",
		Description: description,
		Required:    required,
		Enum:        enum,
	}
	return d
}

// ArrayParam adds an array parameter to the tool.
func (d *ToolDecorator) ArrayParam(name, description string, required bool, itemType string) *ToolDecorator {
	d.params[name] = SimpleParam{
		Type:        "array",
		Description: description,
		Required:    required,
		Items: &SimpleParam{
			Type: itemType,
		},
	}
	return d
}

// ObjectParam adds an object parameter to the tool.
func (d *ToolDecorator) ObjectParam(name, description string, required bool, properties map[string]SimpleParam) *ToolDecorator {
	d.params[name] = SimpleParam{
		Type:        "object",
		Description: description,
		Required:    required,
		Properties:  properties,
	}
	return d
}

// Handle sets the handler function and builds the tool.
func (d *ToolDecorator) Handle(handler ToolFunc) (McpTool, error) {
	d.handler = handler

	simpleTool := SimpleTool{
		Name:        d.name,
		Description: d.description,
		Parameters:  d.params,
		Handler:     handler,
	}

	return simpleTool.Build()
}

// MustHandle is like Handle but panics on error.
// Useful for initialization code where errors should be fatal.
func (d *ToolDecorator) MustHandle(handler ToolFunc) McpTool {
	tool, err := d.Handle(handler)
	if err != nil {
		panic(fmt.Sprintf("failed to create tool %s: %v", d.name, err))
	}
	return tool
}

// QuickTool creates a simple tool with minimal configuration.
// This is the most concise way to create a tool, similar to Python's @tool decorator.
//
// Example:
//
//	tool := types.QuickTool("greet", "Greet a user",
//	    map[string]string{"name": "string"},
//	    func(ctx context.Context, args map[string]interface{}) (*types.ToolResult, error) {
//	        name := args["name"].(string)
//	        return types.NewMcpToolResult(
//	            types.TextBlock{Type: "text", Text: fmt.Sprintf("Hello, %s!", name)},
//	        ), nil
//	    },
//	)
func QuickTool(name, description string, params map[string]string, handler ToolFunc) (McpTool, error) {
	simpleParams := make(map[string]SimpleParam)
	for paramName, paramType := range params {
		simpleParams[paramName] = SimpleParam{
			Type:        paramType,
			Description: fmt.Sprintf("%s parameter", paramName),
			Required:    true, // All params required by default in quick mode
		}
	}

	simpleTool := SimpleTool{
		Name:        name,
		Description: description,
		Parameters:  simpleParams,
		Handler:     handler,
	}
	return simpleTool.Build()
}

// MustQuickTool is like QuickTool but panics on error.
func MustQuickTool(name, description string, params map[string]string, handler ToolFunc) McpTool {
	tool, err := QuickTool(name, description, params, handler)
	if err != nil {
		panic(fmt.Sprintf("failed to create tool %s: %v", name, err))
	}
	return tool
}
