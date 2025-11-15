package main

import (
	"context"
	"fmt"
	"log"

	claude "github.com/M1n9X/claude-agent-sdk-go"
	"github.com/M1n9X/claude-agent-sdk-go/types"
)

// MCPPythonCalculator demonstrates the use of SDK MCP servers to create
// calculator tools that run in-process within your Go application.
// This provides better performance and simpler deployment than external MCP servers.
func main() {
	ctx := context.Background()

	// Create calculator tools using the tool builder pattern
	calculatorTools, err := types.NewCalculatorToolkit()
	if err != nil {
		log.Fatalf("Failed to create calculator tools: %v", err)
	}

	// Create an SDK MCP server with the calculator tools
	server := types.CreateToolServer(
		"calculator",    // server name
		"2.0.0",         // server version
		calculatorTools, // list of tools
	)

	// Configure options to use the calculator server with allowed tools
	// Pre-approve all calculator MCP tools so they can be used without permission prompts
	opts := types.NewClaudeAgentOptions().
		WithMcpServers(map[string]interface{}{
			"calc": server,
		}).
		WithAllowedTools(
			"mcp__calc__add",
			"mcp__calc__subtract",
			"mcp__calc__multiply",
			"mcp__calc__divide",
			"mcp__calc__power",
		)

	// Example prompts to demonstrate calculator usage
	prompts := []string{
		"List your tools",
		"Calculate 15 + 27",
		"What is 100 divided by 7?",
		"What is 2 raised to the power of 8?",
		"Calculate (12 + 8) * 3 - 10", // Complex calculation
	}

	fmt.Println("MCP Calculator Example")
	fmt.Println("======================")

	for _, prompt := range prompts {
		fmt.Printf("\n%s\n", prompt)
		fmt.Println("=====================")

		messages, err := claude.Query(ctx, prompt, opts)
		if err != nil {
			log.Printf("Query failed: %v", err)
			continue
		}

		// Process messages from the channel
		for msg := range messages {
			msgType := msg.GetMessageType()

			switch msgType {
			case "assistant":
				if assistantMsg, ok := msg.(*types.AssistantMessage); ok {
					for _, block := range assistantMsg.Content {
						switch b := block.(type) {
						case *types.TextBlock:
							fmt.Printf("Claude: %s\n", b.Text)
						case *types.ToolUseBlock:
							fmt.Printf("Tool Use: %s\n", b.Name)
							fmt.Printf("  Input: %+v\n", b.Input)
						}
					}
				}
			case "user": // This may contain tool results
				if userMsg, ok := msg.(*types.UserMessage); ok {
					// Handle the interface{} content field that could be []ContentBlock or string
					if contentBlocks, ok := userMsg.Content.([]types.ContentBlock); ok {
						for _, block := range contentBlocks {
							if toolResultBlock, ok := block.(*types.ToolResultBlock); ok {
								fmt.Printf("Tool Result: %s\n", toolResultBlock.Content)
							}
						}
					}
				}
			case "result":
				fmt.Println("---")
				fmt.Println("Query completed successfully")
			}
		}
	}
}
