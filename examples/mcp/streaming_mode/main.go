package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	claude "github.com/M1n9X/claude-agent-sdk-go"
	"github.com/M1n9X/claude-agent-sdk-go/types"
)

// StreamingMode demonstrates various patterns for building applications with the Claude SDK streaming interface.
func main() {
	ctx := context.Background()

	// Run basic streaming example
	if err := basicStreamingExample(ctx); err != nil {
		log.Printf("Basic streaming example failed: %v", err)
	}

	// Run multi-turn conversation example
	if err := multiTurnConversationExample(ctx); err != nil {
		log.Printf("Multi-turn conversation example failed: %v", err)
	}

	// Run concurrent responses example
	if err := concurrentResponsesExample(ctx); err != nil {
		log.Printf("Concurrent responses example failed: %v", err)
	}

	// Run with interrupt example
	if err := withInterruptExample(ctx); err != nil {
		log.Printf("Interrupt example failed: %v", err)
	}

	// Run with options example
	if err := withOptionsExample(ctx); err != nil {
		log.Printf("Options example failed: %v", err)
	}
}

func basicStreamingExample(ctx context.Context) error {
	fmt.Println("=== Basic Streaming Example ===")

	client, err := claude.NewClient(ctx, types.NewClaudeAgentOptions())
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	if err := client.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer client.Close(ctx)

	fmt.Println("User: What is 2+2?")
	if err := client.Query(ctx, "What is 2+2?"); err != nil {
		return fmt.Errorf("query failed: %w", err)
	}

	// Receive complete response
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
	return nil
}

func multiTurnConversationExample(ctx context.Context) error {
	fmt.Println("=== Multi-Turn Conversation Example ===")

	client, err := claude.NewClient(ctx, types.NewClaudeAgentOptions())
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	if err := client.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer client.Close(ctx)

	// First turn
	fmt.Println("User: What's the capital of France?")
	if err := client.Query(ctx, "What's the capital of France?"); err != nil {
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

	// Second turn - follow-up
	fmt.Println("\nUser: What's the population of that city?")
	if err := client.Query(ctx, "What's the population of that city?"); err != nil {
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
	return nil
}

func concurrentResponsesExample(ctx context.Context) error {
	fmt.Println("=== Concurrent Send/Receive Example ===")

	client, err := claude.NewClient(ctx, types.NewClaudeAgentOptions())
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	if err := client.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer client.Close(ctx)

	// Background goroutine to continuously receive messages
	go func() {
		for msg := range client.ReceiveResponse(ctx) {
			if assistantMsg, ok := msg.(*types.AssistantMessage); ok {
				for _, block := range assistantMsg.Content {
					if textBlock, ok := block.(*types.TextBlock); ok {
						fmt.Printf("Claude: %s\n", textBlock.Text)
					}
				}
			}
		}
	}()

	// Send multiple messages with delays
	questions := []string{
		"What is 2 + 2?",
		"What is the square root of 144?",
		"What is 10% of 80?",
	}

	for _, question := range questions {
		fmt.Printf("\nUser: %s\n", question)
		if err := client.Query(ctx, question); err != nil {
			return fmt.Errorf("query failed: %w", err)
		}
		time.Sleep(3 * time.Second) // Wait between messages
	}

	// Give time for final responses
	time.Sleep(2 * time.Second)

	fmt.Println()
	return nil
}

func withInterruptExample(ctx context.Context) error {
	fmt.Println("=== Interrupt Example ===")
	fmt.Println("IMPORTANT: Interrupts require active message consumption.")

	client, err := claude.NewClient(ctx, types.NewClaudeAgentOptions())
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	if err := client.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer client.Close(ctx)

	// Start a long-running task
	fmt.Println("\nUser: Count from 1 to 20 slowly")
	if err := client.Query(ctx, "Count from 1 to 20 slowly, with a brief pause between each number"); err != nil {
		return fmt.Errorf("query failed: %w", err)
	}

	// Start receiving messages in the background to enable interrupt
	done := make(chan bool)
	go func() {
		for msg := range client.ReceiveResponse(ctx) {
			if assistantMsg, ok := msg.(*types.AssistantMessage); ok {
				for _, block := range assistantMsg.Content {
					if textBlock, ok := block.(*types.TextBlock); ok {
						fmt.Printf("Claude: %s\n", textBlock.Text)
					}
				}
			}
		}
		done <- true
	}()

	// Wait 2 seconds then send interrupt
	time.Sleep(2 * time.Second)
	fmt.Println("\n[After 2 seconds, sending interrupt...]")
	if err := client.Interrupt(ctx); err != nil {
		return fmt.Errorf("interrupt failed: %w", err)
	}

	// Wait for the background task to finish
	<-done

	// Send new instruction after interrupt
	fmt.Println("\nUser: Never mind, just tell me a quick joke")
	if err := client.Query(ctx, "Never mind, just tell me a quick joke"); err != nil {
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
	return nil
}

func withOptionsExample(ctx context.Context) error {
	fmt.Println("=== Custom Options Example ===")

	// Configure options
	opts := types.NewClaudeAgentOptions().
		WithAllowedTools("Read", "Write").
		WithSystemPrompt("You are a helpful coding assistant.").
		WithEnv(map[string]string{
			"ANTHROPIC_MODEL": "claude-sonnet-4-5",
		})

	client, err := claude.NewClient(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	if err := client.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer client.Close(ctx)

	fmt.Println("User: Create a simple hello.txt file with a greeting message")
	if err := client.Query(ctx, "Create a simple hello.txt file with a greeting message"); err != nil {
		return fmt.Errorf("query failed: %w", err)
	}

	toolUses := []string{}
	for msg := range client.ReceiveResponse(ctx) {
		if assistantMsg, ok := msg.(*types.AssistantMessage); ok {
			for _, block := range assistantMsg.Content {
				if textBlock, ok := block.(*types.TextBlock); ok {
					fmt.Printf("Claude: %s\n", textBlock.Text)
				}
				if toolUseBlock, ok := block.(*types.ToolUseBlock); ok {
					toolUses = append(toolUses, toolUseBlock.Name)
				}
			}
		}
	}

	if len(toolUses) > 0 {
		fmt.Printf("Tools used: %s\n", strings.Join(toolUses, ", "))
	}

	fmt.Println()
	return nil
}
