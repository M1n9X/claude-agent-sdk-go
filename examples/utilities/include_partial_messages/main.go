package main

import (
	"context"
	"fmt"
	"log"

	claude "github.com/M1n9X/claude-agent-sdk-go"
	"github.com/M1n9X/claude-agent-sdk-go/types"
)

// IncludePartialMessages demonstrates how to configure Claude to return partial
// messages as they are generated, allowing for real-time streaming of responses.
func main() {
	ctx := context.Background()

	fmt.Println("Include Partial Messages Example")
	fmt.Println("===============================")
	fmt.Println("This example shows the difference between regular and partial message streaming.\n")

	// Example 1: Regular streaming (only complete messages)
	fmt.Println("=== Example 1: Regular Streaming (Complete Messages Only) ===")
	opts1 := types.NewClaudeAgentOptions().
		WithModel("claude-sonnet-4-5-20250929")

	fmt.Println("Query: Count from 1 to 10")
	messages1, err := claude.Query(ctx, "Count from 1 to 10, saying each number slowly", opts1)
	if err != nil {
		log.Printf("Query failed: %v", err)
	} else {
		for msg := range messages1 {
			if msg.GetMessageType() == "assistant" {
				if assistantMsg, ok := msg.(*types.AssistantMessage); ok {
					for _, block := range assistantMsg.Content {
						if textBlock, ok := block.(*types.TextBlock); ok {
							fmt.Printf("Complete message: %s\n", textBlock.Text)
						}
					}
				}
			}
		}
	}

	fmt.Println("\n" + "==================================================" + "\n")

	// Example 2: Include partial messages
	fmt.Println("=== Example 2: With Partial Messages ===")
	opts2 := types.NewClaudeAgentOptions().
		// WithModel("claude-sonnet-4-5-20250929").
		WithIncludePartialMessages(true) // Enable partial message streaming

	fmt.Println("Query: Count from 1 to 10 (with partial messages)")
	messages2, err := claude.Query(ctx, "Count from 1 to 10, saying each number slowly", opts2)
	if err != nil {
		log.Printf("Query failed: %v", err)
	} else {
		partialCount := 0
		for msg := range messages2 {
			if msg.GetMessageType() == "assistant" {
				if assistantMsg, ok := msg.(*types.AssistantMessage); ok {
					for _, block := range assistantMsg.Content {
						if textBlock, ok := block.(*types.TextBlock); ok {
							partialCount++
							fmt.Printf("Partial message %d: %s\n", partialCount, textBlock.Text)
						}
					}
				}
			}
		}
		fmt.Printf("\nTotal partial messages received: %d\n", partialCount)
	}

	fmt.Println("\n" + "==================================================" + "\n")

	// Example 3: Use case - Real-time text processing
	fmt.Println("=== Example 3: Real-time Processing Use Case ===")
	fmt.Println("Partial messages are useful for:")
	fmt.Println("- Real-time text display (typing simulation)")
	fmt.Println("- Immediate content analysis")
	fmt.Println("- Responsive UI updates")
	fmt.Println("- Early stopping based on partial content")

	opts3 := types.NewClaudeAgentOptions().
		WithIncludePartialMessages(true)

	messages3, err := claude.Query(ctx, "Write a short story about a robot learning to paint", opts3)
	if err != nil {
		log.Printf("Query failed: %v", err)
	} else {
		wordCount := 0
		for msg := range messages3 {
			if msg.GetMessageType() == "assistant" {
				if assistantMsg, ok := msg.(*types.AssistantMessage); ok {
					for _, block := range assistantMsg.Content {
						if textBlock, ok := block.(*types.TextBlock); ok {
							// Count words in partial message
							words := len([]rune(textBlock.Text)) // Simplified - counts characters
							wordCount += words
							fmt.Printf("[Partial: %d chars so far] %s\n", wordCount, textBlock.Text)
						}
					}
				}
			}
		}
		fmt.Printf("\nTotal character count: %d\n", wordCount)
	}
}
