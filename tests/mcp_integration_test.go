//go:build integration
// +build integration

package tests

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/M1n9X/claude-agent-sdk-go/internal/mcp"
	"github.com/M1n9X/claude-agent-sdk-go/types"
)

// TestSdkMCPFullFlow tests the complete SDK MCP flow
func TestSdkMCPFullFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	_, cancel := CreateTestContext(t, 30*time.Second)
	defer cancel()

	// Create tools
	addTool, err := types.NewTool("add").
		Description("Add two numbers").
		NumberParam("a", "First number", true).
		NumberParam("b", "Second number", true).
		Handler(func(ctx context.Context, args map[string]interface{}) (*types.ToolResult, error) {
			a := args["a"].(float64)
			b := args["b"].(float64)
			sum := a + b
			return types.NewMcpToolResult(
				types.TextBlock{
					Type: "text",
					Text: fmt.Sprintf("%.2f + %.2f = %.2f", a, b, sum),
				},
			), nil
		}).
		Build()

	if err != nil {
		t.Fatalf("Failed to build tool: %v", err)
	}

	// Create server
	server := mcp.NewSdkMCPServer("calculator", "1.0.0", []types.McpTool{addTool})

	// Test initialize
	initMsg := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params":  map[string]interface{}{},
	}

	response, err := server.HandleMessage(initMsg)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Verify response
	if response["error"] != nil {
		t.Errorf("Initialize returned error: %v", response["error"])
	}

	if response["result"] == nil {
		t.Errorf("Expected result in response, got: %v", response)
	}

	result := response["result"].(map[string]interface{})
	if result["protocolVersion"] != "0.1.0" {
		t.Errorf("Unexpected protocol version: %v", result["protocolVersion"])
	}

	// Test tools/list
	listMsg := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/list",
		"params":  map[string]interface{}{},
	}

	response, err = server.HandleMessage(listMsg)
	if err != nil {
		t.Fatalf("Tools list failed: %v", err)
	}

	// Verify tools
	result = response["result"].(map[string]interface{})
	tools, ok := result["tools"].([]interface{})
	if !ok {
		// Handle the case where JSON unmarshaling results in []map[string]interface{}
		if toolsMap, ok2 := result["tools"].([]map[string]interface{}); ok2 {
			// Convert to []interface{} for consistent processing
			tools = make([]interface{}, len(toolsMap))
			for i, v := range toolsMap {
				tools[i] = v
			}
		} else {
			t.Fatalf("Could not convert tools to []interface{} or []map[string]interface{}")
		}
	}
	if len(tools) != 1 {
		t.Errorf("Expected 1 tool, got %d", len(tools))
	}

	tool := tools[0].(map[string]interface{})
	if tool["name"] != "add" {
		t.Errorf("Expected tool name 'add', got: %v", tool["name"])
	}

	// Test tools/call
	callMsg := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      3,
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name": "add",
			"arguments": map[string]interface{}{
				"a": 5.0,
				"b": 3.0,
			},
		},
	}

	response, err = server.HandleMessage(callMsg)
	if err != nil {
		t.Fatalf("Tool call failed: %v", err)
	}

	// Verify result
	if response["error"] != nil {
		t.Errorf("Tool execution returned error: %v", response["error"])
	}

	result = response["result"].(map[string]interface{})
	content, ok := result["content"].([]interface{})
	if !ok {
		// Handle the case where JSON unmarshaling results in []map[string]interface{}
		if contentMap, ok2 := result["content"].([]map[string]interface{}); ok2 {
			content = make([]interface{}, len(contentMap))
			for i, v := range contentMap {
				content[i] = v
			}
		} else {
			t.Fatalf("Could not convert content to []interface{} or []map[string]interface{}")
		}
	}
	textBlock := content[0].(map[string]interface{})
	if !strings.Contains(textBlock["text"].(string), "8.00") {
		t.Errorf("Unexpected result text: %s", textBlock["text"])
	}

	// Test error case - tool not found
	callMsg["params"] = map[string]interface{}{
		"name": "nonexistent",
		"arguments": map[string]interface{}{
			"a": 1.0,
			"b": 2.0,
		},
	}
	callMsg["id"] = 4

	response, err = server.HandleMessage(callMsg)
	if err != nil {
		t.Fatalf("Tool call should not error: %v", err)
	}

	if response["error"] == nil {
		t.Errorf("Expected error for nonexistent tool, got nil")
	}

	errObj := response["error"].(map[string]interface{})
	code, ok := errObj["code"].(float64)
	if !ok {
		// Handle case where code might be int after JSON unmarshaling in different contexts
		if codeInt, ok2 := errObj["code"].(int); ok2 {
			code = float64(codeInt)
		} else if codeInt64, ok3 := errObj["code"].(int64); ok3 {
			code = float64(codeInt64)
		} else {
			t.Fatalf("Could not convert error code to float64 or int")
		}
	}
	if code != -32601 { // ErrorCodeMethodNotFound
		t.Errorf("Expected MethodNotFound error code, got: %v", errObj["code"])
	}
}

// TestToolBuilder tests the ToolBuilder with all parameter types
func TestToolBuilder(t *testing.T) {
	ctx, cancel := CreateTestContext(t, 30*time.Second)
	defer cancel()

	// Test with all parameter types
	tool, err := types.NewTool("complex").
		Description("Complex tool with all parameter types").
		StringParam("str", "String param", true).
		NumberParam("num", "Number param", true).
		IntParam("int", "Integer param", false).
		BoolParam("bool", "Boolean param", false).
		ArrayParam("arr", "Array param", false, "string").
		EnumParam("enum", "Enum param", false, []interface{}{"a", "b", "c"}).
		Handler(func(ctx context.Context, args map[string]interface{}) (*types.ToolResult, error) {
			str := args["str"].(string)
			num := args["num"].(float64)

			result := fmt.Sprintf("str=%s, num=%.2f", str, num)

			// Optional fields
			if val, ok := args["int"]; ok {
				result += fmt.Sprintf(", int=%d", int(val.(float64)))
			}
			if val, ok := args["bool"]; ok {
				result += fmt.Sprintf(", bool=%v", val.(bool))
			}
			if val, ok := args["arr"]; ok {
				result += fmt.Sprintf(", arr=%v", val)
			}
			if val, ok := args["enum"]; ok {
				result += fmt.Sprintf(", enum=%v", val)
			}

			return types.NewMcpToolResult(
				types.TextBlock{
					Type: "text",
					Text: result,
				},
			), nil
		}).
		Build()

	if err != nil {
		t.Fatalf("Failed to build tool: %v", err)
	}

	// Verify schema
	schema := tool.InputSchema()
	if schema["type"] != "object" {
		t.Errorf("Expected object type, got: %v", schema["type"])
	}

	properties := schema["properties"].(map[string]interface{})
	if len(properties) != 6 {
		t.Errorf("Expected 6 properties, got %d", len(properties))
	}

	// Verify parameter types
	paramTypes := map[string]string{
		"str":  "string",
		"num":  "number",
		"int":  "integer",
		"bool": "boolean",
		"arr":  "array",
		"enum": "string",
	}

	for name, expectedType := range paramTypes {
		prop, ok := properties[name].(map[string]interface{})
		if !ok {
			t.Errorf("Property %s not found", name)
			continue
		}

		if prop["type"] != expectedType {
			t.Errorf("Expected %s type for %s, got: %v", expectedType, name, prop["type"])
		}
	}

	// Verify enum values
	enumProp := properties["enum"].(map[string]interface{})
	enumValues := enumProp["enum"].([]interface{})
	if len(enumValues) != 3 {
		t.Errorf("Expected 3 enum values, got %d", len(enumValues))
	}

	// Verify required fields
	required := schema["required"].([]interface{})
	if len(required) != 2 {
		t.Errorf("Expected 2 required fields, got %d", len(required))
	}

	// Test execution
	input := map[string]interface{}{
		"str":  "test",
		"num":  42.0,
		"bool": true,
		"arr":  []interface{}{"a", "b"},
		"enum": "b",
	}

	result, err := tool.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Tool execution failed: %v", err)
	}

	content := result.Content[0].(types.TextBlock).Text
	if !strings.Contains(content, "str=test") || !strings.Contains(content, "num=42.00") {
		t.Errorf("Unexpected result: %s", content)
	}
}

// TestToolValidation tests parameter validation
func TestToolValidation(t *testing.T) {
	ctx, cancel := CreateTestContext(t, 30*time.Second)
	defer cancel()

	// Create a tool with validation
	tool, err := types.NewTool("validated").
		Description("Tool with custom validation").
		StringParam("email", "Email address", true).
		WithValidation(func(input map[string]interface{}) error {
			email, ok := input["email"].(string)
			if !ok {
				return fmt.Errorf("email must be a string")
			}
			// Simple email validation
			if !strings.Contains(email, "@") {
				return fmt.Errorf("invalid email format")
			}
			return nil
		}).
		Handler(func(ctx context.Context, args map[string]interface{}) (*types.ToolResult, error) {
			email := args["email"].(string)
			return types.NewMcpToolResult(
				types.TextBlock{
					Type: "text",
					Text: fmt.Sprintf("Email validated: %s", email),
				},
			), nil
		}).
		Build()

	if err != nil {
		t.Fatalf("Failed to build tool: %v", err)
	}

	// Test valid input
	validInput := map[string]interface{}{
		"email": "test@example.com",
	}

	result, err := tool.Execute(ctx, validInput)
	if err != nil {
		t.Fatalf("Valid input should not error: %v", err)
	}

	if result.IsError {
		t.Error("Valid input should not return error result")
	}

	// Test invalid email format
	invalidInput := map[string]interface{}{
		"email": "notanemail",
	}

	_, err = tool.Execute(ctx, invalidInput)
	if err == nil {
		t.Error("Invalid email format should error")
	}

	if !strings.Contains(err.Error(), "invalid email format") {
		t.Errorf("Expected validation error, got: %v", err)
	}

	// Test missing required field
	missingInput := map[string]interface{}{}

	_, err = tool.Execute(ctx, missingInput)
	if err == nil {
		t.Error("Missing required field should error")
	}
}

// TestToolManager tests the ToolManager
func TestToolManager(t *testing.T) {
	_, cancel := CreateTestContext(t, 30*time.Second)
	defer cancel()

	// Create tool manager
	manager := types.NewToolManager()

	// Create tools
	tool1, err := types.NewTool("tool1").
		Description("First tool").
		StringParam("input", "Input", true).
		Handler(func(ctx context.Context, args map[string]interface{}) (*types.ToolResult, error) {
			return types.NewMcpToolResult(
				types.TextBlock{
					Type: "text",
					Text: "tool1 executed",
				},
			), nil
		}).
		Build()

	if err != nil {
		t.Fatalf("Failed to build tool: %v", err)
	}

	tool2, err := types.NewTool("tool2").
		Description("Second tool").
		StringParam("input", "Input", true).
		Handler(func(ctx context.Context, args map[string]interface{}) (*types.ToolResult, error) {
			return types.NewMcpToolResult(
				types.TextBlock{
					Type: "text",
					Text: "tool2 executed",
				},
			), nil
		}).
		Build()

	if err != nil {
		t.Fatalf("Failed to build tool: %v", err)
	}

	// Register tools
	if err := manager.Register(tool1); err != nil {
		t.Fatalf("Failed to register tool1: %v", err)
	}

	if err := manager.Register(tool2); err != nil {
		t.Fatalf("Failed to register tool2: %v", err)
	}

	// Test duplicate registration
	if err := manager.Register(tool1); err == nil {
		t.Error("Duplicate registration should error")
	}

	// Test Get
	retrieved, exists := manager.Get("tool1")
	if !exists {
		t.Error("tool1 should exist")
	}

	if retrieved.Name() != "tool1" {
		t.Errorf("Expected tool name 'tool1', got: %s", retrieved.Name())
	}

	// Test List
	tools := manager.List()
	if len(tools) != 2 {
		t.Errorf("Expected 2 tools, got %d", len(tools))
	}

	// Test CreateServer
	server := manager.CreateServer("testapp", "1.0.0")
	if server.Name != "testapp" {
		t.Errorf("Expected server name 'testapp', got: %s", server.Name)
	}

	// Verify server has both tools
	toolList := server.Instance.(*mcp.SdkMCPServer).Tools()
	if len(toolList) != 2 {
		t.Errorf("Expected 2 tools in server, got %d", len(toolList))
	}
}

// TestObjectArrayParam tests object array parameters
func TestObjectArrayParam(t *testing.T) {
	ctx, cancel := CreateTestContext(t, 30*time.Second)
	defer cancel()

	// Create item schema
	itemSchema := map[string]types.ToolParam{
		"id":    {Type: "string", Description: "Item ID", Required: true},
		"name":  {Type: "string", Description: "Item name", Required: true},
		"price": {Type: "number", Description: "Item price"},
	}

	tool, err := types.NewTool("process_items").
		Description("Process a list of items").
		ObjectArrayParam("items", "List of items", true, itemSchema).
		Handler(func(ctx context.Context, args map[string]interface{}) (*types.ToolResult, error) {
			items := args["items"].([]interface{})
			total := 0.0

			for _, item := range items {
				itemMap := item.(map[string]interface{})
				price, ok := itemMap["price"].(float64)
				if ok {
					total += price
				}
			}

			return types.NewMcpToolResult(
				types.TextBlock{
					Type: "text",
					Text: fmt.Sprintf("Processed %d items, total price: %.2f", len(items), total),
				},
			), nil
		}).
		Build()

	if err != nil {
		t.Fatalf("Failed to build tool: %v", err)
	}

	// Verify schema
	schema := tool.InputSchema()
	properties := schema["properties"].(map[string]interface{})
	itemsProp := properties["items"].(map[string]interface{})

	if itemsProp["type"] != "array" {
		t.Errorf("Expected array type, got: %v", itemsProp["type"])
	}

	itemsSchema := itemsProp["items"].(map[string]interface{})
	if itemsSchema["type"] != "object" {
		t.Errorf("Expected object type for items, got: %v", itemsSchema["type"])
	}

	itemProps := itemsSchema["properties"].(map[string]interface{})
	if len(itemProps) != 3 {
		t.Errorf("Expected 3 item properties, got %d", len(itemProps))
	}

	// Test execution
	input := map[string]interface{}{
		"items": []interface{}{
			map[string]interface{}{
				"id":    "item1",
				"name":  "Item 1",
				"price": 10.0,
			},
			map[string]interface{}{
				"id":    "item2",
				"name":  "Item 2",
				"price": 20.0,
			},
		},
	}

	result, err := tool.Execute(ctx, input)
	if err != nil {
		t.Fatalf("Tool execution failed: %v", err)
	}

	content := result.Content[0].(types.TextBlock).Text
	if !strings.Contains(content, "processed 2 items") || !strings.Contains(content, "total price: 30.00") {
		t.Errorf("Unexpected result: %s", content)
	}
}
