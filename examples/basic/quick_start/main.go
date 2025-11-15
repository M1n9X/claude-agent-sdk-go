package main

import (
	"context"
	"fmt"
	"log"

	claude "github.com/M1n9X/claude-agent-sdk-go"
	"github.com/M1n9X/claude-agent-sdk-go/types"
)

// QuickStart demonstrates a quick start example for Claude Code SDK.
func main() {
	ctx := context.Background()

	// Run basic example
	if err := basicExample(ctx); err != nil {
		log.Printf("Basic example failed: %v", err)
	}

	// Run with options example
	if err := withOptionsExample(ctx); err != nil {
		log.Printf("With options example failed: %v", err)
	}

	// Run with tools example
	if err := withToolsExample(ctx); err != nil {
		log.Printf("With tools example failed: %v", err)
	}
}

func basicExample(ctx context.Context) error {
	fmt.Println("=== Basic Example ===")

	// Simple query
	messages, err := claude.Query(ctx, "What is 2 + 2?", types.NewClaudeAgentOptions())
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}

	for msg := range messages {
		if assistantMsg, ok := msg.(*types.AssistantMessage); ok {
			for _, block := range assistantMsg.Content {
				if textBlock, ok := block.(*types.TextBlock); ok {
					fmt.Printf("Claude: %s\n", textBlock.Text)
				}
			}
		}
	}

	fmt.Println()
	return nil
}

func withOptionsExample(ctx context.Context) error {
	fmt.Println("=== With Options Example ===")

	// Create options with system prompt and max turns
	opts := types.NewClaudeAgentOptions().
		WithSystemPrompt("You are a helpful assistant that explains things simply.").
		WithMaxTurns(1)

	messages, err := claude.Query(ctx, "Explain what Go is in one sentence.", opts)
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}

	for msg := range messages {
		if assistantMsg, ok := msg.(*types.AssistantMessage); ok {
			for _, block := range assistantMsg.Content {
				if textBlock, ok := block.(*types.TextBlock); ok {
					fmt.Printf("Claude: %s\n", textBlock.Text)
				}
			}
		}
	}

	fmt.Println()
	return nil
}

func withToolsExample(ctx context.Context) error {
	fmt.Println("=== With Tools Example ===")

	// Create options with allowed tools
	opts := types.NewClaudeAgentOptions().
		WithAllowedTools("Read", "Write").
		WithSystemPrompt("You are a helpful file assistant.")

	messages, err := claude.Query(ctx, "Create a file called hello.txt with 'Hello, World!' in it", opts)
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}

	for msg := range messages {
		if assistantMsg, ok := msg.(*types.AssistantMessage); ok {
			for _, block := range assistantMsg.Content {
				if textBlock, ok := block.(*types.TextBlock); ok {
					fmt.Printf("Claude: %s\n", textBlock.Text)
				}
			}
		} else if resultMsg, ok := msg.(*types.ResultMessage); ok && resultMsg.TotalCostUSD != nil && *resultMsg.TotalCostUSD > 0 {
			fmt.Printf("\nCost: $%.4f\n", *resultMsg.TotalCostUSD)
		}
	}

	fmt.Println()
	return nil
}
