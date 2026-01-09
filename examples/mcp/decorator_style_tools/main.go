package main

import (
	"context"
	"fmt"
	"log"

	claude "github.com/M1n9X/claude-agent-sdk-go"
	"github.com/M1n9X/claude-agent-sdk-go/types"
)

// DecoratorStyleTools demonstrates the decorator-style API for creating tools,
// similar to Python's @tool decorator. This provides a more concise way to
// define tools compared to the builder pattern.
func main() {
	ctx := context.Background()

	// Method 1: SimpleTool struct (most similar to Python's @tool decorator)
	greetTool := types.SimpleTool{
		Name:        "greet",
		Description: "Greet a user by name",
		Parameters: map[string]types.SimpleParam{
			"name": {
				Type:        "string",
				Description: "The name of the person to greet",
				Required:    true,
			},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (*types.ToolResult, error) {
			name := args["name"].(string)
			return types.NewMcpToolResult(
				types.TextBlock{
					Type: "text",
					Text: fmt.Sprintf("Hello, %s! Welcome to the decorator-style tools example!", name),
				},
			), nil
		},
	}

	greetToolBuilt, err := greetTool.Build()
	if err != nil {
		log.Fatalf("Failed to build greet tool: %v", err)
	}

	// Method 2: Fluent API with Tool() function
	calculateTool, err := types.Tool("calculate", "Perform a calculation").
		Param("operation", "string", "The operation to perform", true).
		EnumParam("operation", "The operation to perform", true, []interface{}{"add", "subtract", "multiply", "divide"}).
		Param("a", "number", "First number", true).
		Param("b", "number", "Second number", true).
		Handle(func(ctx context.Context, args map[string]interface{}) (*types.ToolResult, error) {
			operation := args["operation"].(string)
			a := args["a"].(float64)
			b := args["b"].(float64)

			var result float64
			var resultText string

			switch operation {
			case "add":
				result = a + b
				resultText = fmt.Sprintf("%.2f + %.2f = %.2f", a, b, result)
			case "subtract":
				result = a - b
				resultText = fmt.Sprintf("%.2f - %.2f = %.2f", a, b, result)
			case "multiply":
				result = a * b
				resultText = fmt.Sprintf("%.2f × %.2f = %.2f", a, b, result)
			case "divide":
				if b == 0 {
					return types.NewErrorMcpToolResult("Cannot divide by zero"), nil
				}
				result = a / b
				resultText = fmt.Sprintf("%.2f ÷ %.2f = %.2f", a, b, result)
			default:
				return types.NewErrorMcpToolResult(fmt.Sprintf("Unknown operation: %s", operation)), nil
			}

			return types.NewMcpToolResult(
				types.TextBlock{
					Type: "text",
					Text: resultText,
				},
			), nil
		})
	if err != nil {
		log.Fatalf("Failed to build calculate tool: %v", err)
	}

	// Method 3: QuickTool for ultra-simple tools
	echoTool, err := types.QuickTool(
		"echo",
		"Echo back the input message",
		map[string]string{"message": "string"},
		func(ctx context.Context, args map[string]interface{}) (*types.ToolResult, error) {
			message := args["message"].(string)
			return types.NewMcpToolResult(
				types.TextBlock{
					Type: "text",
					Text: fmt.Sprintf("Echo: %s", message),
				},
			), nil
		},
	)
	if err != nil {
		log.Fatalf("Failed to build echo tool: %v", err)
	}

	// Method 4: Complex tool with nested objects
	userInfoTool := types.SimpleTool{
		Name:        "create_user",
		Description: "Create a user profile with detailed information",
		Parameters: map[string]types.SimpleParam{
			"name": {
				Type:        "string",
				Description: "User's full name",
				Required:    true,
			},
			"age": {
				Type:        "integer",
				Description: "User's age",
				Required:    true,
			},
			"address": {
				Type:        "object",
				Description: "User's address",
				Required:    false,
				Properties: map[string]types.SimpleParam{
					"street": {
						Type:        "string",
						Description: "Street address",
						Required:    true,
					},
					"city": {
						Type:        "string",
						Description: "City",
						Required:    true,
					},
					"country": {
						Type:        "string",
						Description: "Country",
						Required:    true,
					},
				},
			},
			"hobbies": {
				Type:        "array",
				Description: "List of hobbies",
				Required:    false,
				Items: &types.SimpleParam{
					Type: "string",
				},
			},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (*types.ToolResult, error) {
			name := args["name"].(string)
			age := int(args["age"].(float64))

			result := fmt.Sprintf("Created user profile:\nName: %s\nAge: %d\n", name, age)

			if address, ok := args["address"].(map[string]interface{}); ok {
				result += fmt.Sprintf("Address: %s, %s, %s\n",
					address["street"], address["city"], address["country"])
			}

			if hobbies, ok := args["hobbies"].([]interface{}); ok {
				result += "Hobbies: "
				for i, hobby := range hobbies {
					if i > 0 {
						result += ", "
					}
					result += hobby.(string)
				}
				result += "\n"
			}

			return types.NewMcpToolResult(
				types.TextBlock{
					Type: "text",
					Text: result,
				},
			), nil
		},
	}

	userInfoToolBuilt, err := userInfoTool.Build()
	if err != nil {
		log.Fatalf("Failed to build user info tool: %v", err)
	}

	// Create an SDK MCP server with all tools
	tools := []types.McpTool{
		greetToolBuilt,
		calculateTool,
		echoTool,
		userInfoToolBuilt,
	}

	server := types.CreateToolServer("decorator-tools", "1.0.0", tools)

	// Configure options
	opts := types.NewClaudeAgentOptions().
		WithMcpServers(map[string]interface{}{
			"tools": server,
		}).
		WithAllowedTools(
			"mcp__tools__greet",
			"mcp__tools__calculate",
			"mcp__tools__echo",
			"mcp__tools__create_user",
		)

	// Example prompts
	prompts := []string{
		"List your available tools",
		"Greet Alice",
		"Calculate 15 + 27",
		"Echo the message 'Hello from decorator-style tools!'",
		"Create a user profile for John Doe, age 30, living at 123 Main St, New York, USA, with hobbies: reading, coding, hiking",
	}

	fmt.Println("Decorator-Style Tools Example")
	fmt.Println("==============================")
	fmt.Println("\nThis example demonstrates three ways to create tools:")
	fmt.Println("1. SimpleTool struct (most similar to Python's @tool)")
	fmt.Println("2. Fluent API with Tool() function")
	fmt.Println("3. QuickTool for ultra-simple tools")
	fmt.Println("4. Complex tools with nested objects and arrays")
	fmt.Println()

	for i, prompt := range prompts {
		fmt.Printf("\n[%d] %s\n", i+1, prompt)
		fmt.Println("------------------------------------------------------------")

		messages, err := claude.Query(ctx, prompt, opts)
		if err != nil {
			log.Printf("Query failed: %v", err)
			continue
		}

		for msg := range messages {
			switch m := msg.(type) {
			case *types.AssistantMessage:
				for _, block := range m.Content {
					if tb, ok := block.(*types.TextBlock); ok {
						fmt.Printf("Claude: %s\n", tb.Text)
					}
				}
			case *types.ResultMessage:
				if m.TotalCostUSD != nil && *m.TotalCostUSD > 0 {
					fmt.Printf("\n💰 Cost: $%.4f\n", *m.TotalCostUSD)
				}
			}
		}
	}

	fmt.Println("\n============================================================")
	fmt.Println("Example completed!")
}
