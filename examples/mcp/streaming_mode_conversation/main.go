package main

import (
	"context"
	"fmt"
	"log"

	claude "github.com/M1n9X/claude-agent-sdk-go"
	"github.com/M1n9X/claude-agent-sdk-go/types"
)

// StreamingModeConversation demonstrates multi-turn conversation patterns using goroutines
// This shows how to use the Claude SDK Client with goroutines for interactive,
// stateful conversations where you can send follow-up messages based on Claude's responses.
func main() {
	ctx := context.Background()

	if err := multiTurnConversation(ctx); err != nil {
		log.Printf("Multi-turn conversation example failed: %v", err)
	}
}

func multiTurnConversation(ctx context.Context) error {
	fmt.Println("=== Multi-turn Conversation with Goroutines ===")

	client, err := claude.NewClient(ctx, types.NewClaudeAgentOptions())
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	if err := client.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer client.Close(ctx)

	// First turn: Simple math question
	fmt.Println("User: What's 15 + 27?")
	if err := client.Query(ctx, "What's 15 + 27?"); err != nil {
		return fmt.Errorf("query failed: %w", err)
	}

	for msg := range client.ReceiveResponse(ctx) {
		if assistantMsg, ok := msg.(*types.AssistantMessage); ok {
			for _, block := range assistantMsg.Content {
				if textBlock, ok := block.(*types.TextBlock); ok {
					fmt.Printf("Claude: %s\n", textBlock.Text)
				}
			}
		}
	}
	fmt.Println()

	// Second turn: Follow-up calculation
	fmt.Println("User: Now multiply that result by 2")
	if err := client.Query(ctx, "Now multiply that result by 2"); err != nil {
		return fmt.Errorf("query failed: %w", err)
	}

	for msg := range client.ReceiveResponse(ctx) {
		if assistantMsg, ok := msg.(*types.AssistantMessage); ok {
			for _, block := range assistantMsg.Content {
				if textBlock, ok := block.(*types.TextBlock); ok {
					fmt.Printf("Claude: %s\n", textBlock.Text)
				}
			}
		}
	}
	fmt.Println()

	// Third turn: One more operation
	fmt.Println("User: Divide that by 7 and round to 2 decimal places")
	if err := client.Query(ctx, "Divide that by 7 and round to 2 decimal places"); err != nil {
		return fmt.Errorf("query failed: %w", err)
	}

	for msg := range client.ReceiveResponse(ctx) {
		if assistantMsg, ok := msg.(*types.AssistantMessage); ok {
			for _, block := range assistantMsg.Content {
				if textBlock, ok := block.(*types.TextBlock); ok {
					fmt.Printf("Claude: %s\n", textBlock.Text)
				}
			}
		}
	}

	fmt.Println("\nConversation complete!")
	return nil
}
