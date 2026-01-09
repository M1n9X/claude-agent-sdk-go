package types

import (
	"context"
	"fmt"
	"testing"
)

func TestSimpleTool(t *testing.T) {
	tool := SimpleTool{
		Name:        "test_tool",
		Description: "A test tool",
		Parameters: map[string]SimpleParam{
			"name": {
				Type:        "string",
				Description: "A name parameter",
				Required:    true,
			},
			"age": {
				Type:        "integer",
				Description: "An age parameter",
				Required:    false,
			},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
			name := args["name"].(string)
			return NewMcpToolResult(
				TextBlock{Type: "text", Text: fmt.Sprintf("Hello, %s!", name)},
			), nil
		},
	}

	mcpTool, err := tool.Build()
	if err != nil {
		t.Fatalf("Failed to build tool: %v", err)
	}

	if mcpTool.Name() != "test_tool" {
		t.Errorf("Expected name 'test_tool', got '%s'", mcpTool.Name())
	}

	if mcpTool.Description() != "A test tool" {
		t.Errorf("Expected description 'A test tool', got '%s'", mcpTool.Description())
	}

	schema := mcpTool.InputSchema()
	if schema["type"] != "object" {
		t.Errorf("Expected schema type 'object', got '%v'", schema["type"])
	}

	properties, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected properties to be a map")
	}

	if _, exists := properties["name"]; !exists {
		t.Error("Expected 'name' property to exist")
	}

	if _, exists := properties["age"]; !exists {
		t.Error("Expected 'age' property to exist")
	}

	required, ok := schema["required"].([]string)
	if !ok {
		t.Fatal("Expected required to be a string slice")
	}

	if len(required) != 1 || required[0] != "name" {
		t.Errorf("Expected required to be ['name'], got %v", required)
	}
}

func TestSimpleToolExecution(t *testing.T) {
	tool := SimpleTool{
		Name:        "greet",
		Description: "Greet a user",
		Parameters: map[string]SimpleParam{
			"name": {Type: "string", Description: "User's name", Required: true},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
			name := args["name"].(string)
			return NewMcpToolResult(
				TextBlock{Type: "text", Text: fmt.Sprintf("Hello, %s!", name)},
			), nil
		},
	}

	mcpTool, err := tool.Build()
	if err != nil {
		t.Fatalf("Failed to build tool: %v", err)
	}

	ctx := context.Background()
	result, err := mcpTool.Execute(ctx, map[string]interface{}{
		"name": "Alice",
	})

	if err != nil {
		t.Fatalf("Tool execution failed: %v", err)
	}

	if result.IsError {
		t.Error("Expected successful result, got error")
	}

	if len(result.Content) != 1 {
		t.Fatalf("Expected 1 content block, got %d", len(result.Content))
	}

	textBlock, ok := result.Content[0].(TextBlock)
	if !ok {
		t.Fatal("Expected TextBlock")
	}

	expected := "Hello, Alice!"
	if textBlock.Text != expected {
		t.Errorf("Expected text '%s', got '%s'", expected, textBlock.Text)
	}
}

func TestToolDecorator(t *testing.T) {
	tool, err := Tool("test", "Test tool").
		Param("name", "string", "Name parameter", true).
		Param("age", "integer", "Age parameter", false).
		Handle(func(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
			name := args["name"].(string)
			return NewMcpToolResult(
				TextBlock{Type: "text", Text: fmt.Sprintf("Hello, %s!", name)},
			), nil
		})

	if err != nil {
		t.Fatalf("Failed to create tool: %v", err)
	}

	if tool.Name() != "test" {
		t.Errorf("Expected name 'test', got '%s'", tool.Name())
	}

	schema := tool.InputSchema()
	properties, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected properties to be a map")
	}

	if len(properties) != 2 {
		t.Errorf("Expected 2 properties, got %d", len(properties))
	}
}

func TestToolDecoratorEnum(t *testing.T) {
	tool, err := Tool("operation", "Perform operation").
		EnumParam("op", "Operation type", true, []interface{}{"add", "subtract"}).
		Param("a", "number", "First number", true).
		Param("b", "number", "Second number", true).
		Handle(func(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
			op := args["op"].(string)
			a := args["a"].(float64)
			b := args["b"].(float64)

			var result float64
			switch op {
			case "add":
				result = a + b
			case "subtract":
				result = a - b
			}

			return NewMcpToolResult(
				TextBlock{Type: "text", Text: fmt.Sprintf("%.2f", result)},
			), nil
		})

	if err != nil {
		t.Fatalf("Failed to create tool: %v", err)
	}

	schema := tool.InputSchema()
	properties, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected properties to be a map")
	}

	opProp, ok := properties["op"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected op property to be a map")
	}

	enum, ok := opProp["enum"].([]interface{})
	if !ok {
		t.Fatal("Expected enum to be a slice")
	}

	if len(enum) != 2 {
		t.Errorf("Expected 2 enum values, got %d", len(enum))
	}
}

func TestQuickTool(t *testing.T) {
	tool, err := QuickTool(
		"echo",
		"Echo a message",
		map[string]string{"message": "string"},
		func(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
			message := args["message"].(string)
			return NewMcpToolResult(
				TextBlock{Type: "text", Text: message},
			), nil
		},
	)

	if err != nil {
		t.Fatalf("Failed to create quick tool: %v", err)
	}

	if tool.Name() != "echo" {
		t.Errorf("Expected name 'echo', got '%s'", tool.Name())
	}

	ctx := context.Background()
	result, err := tool.Execute(ctx, map[string]interface{}{
		"message": "Hello, World!",
	})

	if err != nil {
		t.Fatalf("Tool execution failed: %v", err)
	}

	textBlock, ok := result.Content[0].(TextBlock)
	if !ok {
		t.Fatal("Expected TextBlock")
	}

	if textBlock.Text != "Hello, World!" {
		t.Errorf("Expected 'Hello, World!', got '%s'", textBlock.Text)
	}
}

func TestSimpleToolNestedObject(t *testing.T) {
	tool := SimpleTool{
		Name:        "create_user",
		Description: "Create a user",
		Parameters: map[string]SimpleParam{
			"name": {
				Type:        "string",
				Description: "User's name",
				Required:    true,
			},
			"address": {
				Type:        "object",
				Description: "User's address",
				Required:    false,
				Properties: map[string]SimpleParam{
					"street": {Type: "string", Description: "Street", Required: true},
					"city":   {Type: "string", Description: "City", Required: true},
				},
			},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
			name := args["name"].(string)
			result := fmt.Sprintf("User: %s", name)

			if address, ok := args["address"].(map[string]interface{}); ok {
				result += fmt.Sprintf(", Address: %s, %s",
					address["street"], address["city"])
			}

			return NewMcpToolResult(
				TextBlock{Type: "text", Text: result},
			), nil
		},
	}

	mcpTool, err := tool.Build()
	if err != nil {
		t.Fatalf("Failed to build tool: %v", err)
	}

	schema := mcpTool.InputSchema()
	properties, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected properties to be a map")
	}

	addressProp, ok := properties["address"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected address property to be a map")
	}

	if addressProp["type"] != "object" {
		t.Errorf("Expected address type to be 'object', got '%v'", addressProp["type"])
	}

	addressProps, ok := addressProp["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected address properties to be a map")
	}

	if len(addressProps) != 2 {
		t.Errorf("Expected 2 address properties, got %d", len(addressProps))
	}
}

func TestSimpleToolArray(t *testing.T) {
	tool := SimpleTool{
		Name:        "process_items",
		Description: "Process a list of items",
		Parameters: map[string]SimpleParam{
			"items": {
				Type:        "array",
				Description: "List of items",
				Required:    true,
				Items: &SimpleParam{
					Type: "string",
				},
			},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
			items := args["items"].([]interface{})
			count := len(items)
			return NewMcpToolResult(
				TextBlock{Type: "text", Text: fmt.Sprintf("Processed %d items", count)},
			), nil
		},
	}

	mcpTool, err := tool.Build()
	if err != nil {
		t.Fatalf("Failed to build tool: %v", err)
	}

	schema := mcpTool.InputSchema()
	properties, ok := schema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected properties to be a map")
	}

	itemsProp, ok := properties["items"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected items property to be a map")
	}

	if itemsProp["type"] != "array" {
		t.Errorf("Expected items type to be 'array', got '%v'", itemsProp["type"])
	}

	itemsSchema, ok := itemsProp["items"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected items schema to be a map")
	}

	if itemsSchema["type"] != "string" {
		t.Errorf("Expected items type to be 'string', got '%v'", itemsSchema["type"])
	}
}

func TestSimpleToolValidation(t *testing.T) {
	tests := []struct {
		name        string
		tool        SimpleTool
		expectError bool
		errorMsg    string
	}{
		{
			name: "missing name",
			tool: SimpleTool{
				Description: "Test",
				Handler:     func(ctx context.Context, args map[string]interface{}) (*ToolResult, error) { return nil, nil },
			},
			expectError: true,
			errorMsg:    "tool name is required",
		},
		{
			name: "missing description",
			tool: SimpleTool{
				Name:    "test",
				Handler: func(ctx context.Context, args map[string]interface{}) (*ToolResult, error) { return nil, nil },
			},
			expectError: true,
			errorMsg:    "tool description is required",
		},
		{
			name: "missing handler",
			tool: SimpleTool{
				Name:        "test",
				Description: "Test",
			},
			expectError: true,
			errorMsg:    "tool handler is required",
		},
		{
			name: "valid tool",
			tool: SimpleTool{
				Name:        "test",
				Description: "Test",
				Handler:     func(ctx context.Context, args map[string]interface{}) (*ToolResult, error) { return nil, nil },
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.tool.Build()
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if err.Error() != tt.errorMsg {
					t.Errorf("Expected error '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}
