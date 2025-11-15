package main

import (
	"context"
	"fmt"
	"log"

	claude "github.com/M1n9X/claude-agent-sdk-go"
	"github.com/M1n9X/claude-agent-sdk-go/types"
)

// MaxBudgetUSD demonstrates how to set a maximum budget in USD to control
// costs during Claude agent execution.
func main() {
	ctx := context.Background()

	fmt.Println("Max Budget USD Example")
	fmt.Println("======================")
	fmt.Println("This example shows how to set spending limits to control costs.\n")

	// Example 1: Set a low budget limit
	fmt.Println("=== Example 1: Low Budget Limit ($0.01) ===")
	fmt.Println("Setting a very low budget to demonstrate the limit in action...")

	opts1 := types.NewClaudeAgentOptions().
		// WithModel("claude-sonnet-4-5-20250929").
		WithMaxBudgetUSD(0.01) // Very low budget to trigger limit quickly

	fmt.Println("Query: Write a very long technical document about Go programming")
	messages1, err := claude.Query(ctx, "Write a very long technical document about Go programming with extensive examples", opts1)
	if err != nil {
		log.Printf("Query failed: %v", err)
	} else {
		responseCount := 0
		for msg := range messages1 {
			msgType := msg.GetMessageType()
			if msgType == "assistant" {
				if assistantMsg, ok := msg.(*types.AssistantMessage); ok {
					for _, block := range assistantMsg.Content {
						if textBlock, ok := block.(*types.TextBlock); ok {
							responseCount++
							fmt.Printf("Response %d: %s\n", responseCount, textBlock.Text)
						}
					}
				}
			} else if msgType == "result" {
				if resultMsg, ok := msg.(*types.ResultMessage); ok {
					if resultMsg.TotalCostUSD != nil {
						fmt.Printf("\nTotal cost: $%.6f\n", *resultMsg.TotalCostUSD)
					}
					// Note: Go SDK doesn't have StopReason in ResultMessage like Python SDK
				}
			}
		}
	}

	fmt.Println("\n" + "==================================================" + "\n")

	// Example 2: Set a reasonable budget for testing
	fmt.Println("=== Example 2: Reasonable Budget Limit ($0.10) ===")
	fmt.Println("Setting a small but reasonable budget for a complex task...")

	opts2 := types.NewClaudeAgentOptions().
		// WithModel("claude-sonnet-4-5-20250929").
		WithMaxBudgetUSD(0.10) // Small budget for testing

	fmt.Println("Query: Analyze a complex problem and provide detailed solution")
	messages2, err := claude.Query(ctx, "Analyze the time complexity of various sorting algorithms and explain when to use each one", opts2)
	if err != nil {
		log.Printf("Query failed: %v", err)
	} else {
		for msg := range messages2 {
			if msg.GetMessageType() == "assistant" {
				if assistantMsg, ok := msg.(*types.AssistantMessage); ok {
					for _, block := range assistantMsg.Content {
						if textBlock, ok := block.(*types.TextBlock); ok {
							fmt.Printf("Analysis: %s\n", textBlock.Text)
						}
					}
				}
			} else if msg.GetMessageType() == "result" {
				if resultMsg, ok := msg.(*types.ResultMessage); ok {
					if resultMsg.TotalCostUSD != nil {
						fmt.Printf("\nTotal cost: $%.6f\n", *resultMsg.TotalCostUSD)
					}
					// Note: Go SDK doesn't have BudgetUsedUSD or StopReason in ResultMessage like Python SDK
				}
			}
		}
	}

	fmt.Println("\n" + "==================================================" + "\n")

	// Example 3: Budget with tool usage
	fmt.Println("=== Example 3: Budget with Tool Usage ===")
	fmt.Println("Budget limits also apply when Claude uses tools...")

	opts3 := types.NewClaudeAgentOptions().
		WithMaxBudgetUSD(0.05). // Small budget
		WithAllowedTools("Bash", "Read", "Write")

	fmt.Println("Query: Perform multiple operations that might use tools")
	messages3, err := claude.Query(ctx, "Check current directory, list files, and briefly explain what you see", opts3)
	if err != nil {
		log.Printf("Query failed: %v", err)
	} else {
		for msg := range messages3 {
			msgType := msg.GetMessageType()
			if msgType == "assistant" {
				if assistantMsg, ok := msg.(*types.AssistantMessage); ok {
					for _, block := range assistantMsg.Content {
						switch v := block.(type) {
						case *types.TextBlock:
							fmt.Printf("Claude: %s\n", v.Text)
						case *types.ToolUseBlock:
							fmt.Printf("Tool Use: %s with %v\n", v.Name, v.Input)
						}
					}
				}
			} else if msgType == "user" {
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
			} else if msgType == "result" {
				if resultMsg, ok := msg.(*types.ResultMessage); ok {
					if resultMsg.TotalCostUSD != nil {
						fmt.Printf("\nTotal cost: $%.6f\n", *resultMsg.TotalCostUSD)
					}
					// Note: Go SDK doesn't have BudgetUsedUSD in ResultMessage like Python SDK
				}
			}
		}
	}

	fmt.Println("\n" + "==================================================" + "\n")
	fmt.Println("Budget Controls Summary:")
	fmt.Println("- Use WithMaxBudgetUSD() to set spending limits")
	fmt.Println("- Claude will stop execution when budget is reached")
	fmt.Println("- Cost tracking includes all API usage during the session")
	fmt.Println("- Particularly useful for automated or long-running tasks")
}
