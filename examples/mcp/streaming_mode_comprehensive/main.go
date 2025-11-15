package main

import (
	"context"
	"fmt"
	"log"
	"time"

	claude "github.com/M1n9X/claude-agent-sdk-go"
	"github.com/M1n9X/claude-agent-sdk-go/types"
)

// StreamingModeIPython demonstrates IPython-friendly code snippets for Claude SDK streaming mode.
// These examples are designed to be copy-pasted directly into IPython. Each example is self-contained and can be run independently.
func main() {
	ctx := context.Background()

	// Run basic streaming example
	basicStreaming(ctx)

	// Run streaming with real-time display example
	streamingWithRealTimeDisplay(ctx)

	// Run persistent client example
	persistentClientExample(ctx)

	// Run with interrupt capability example
	withInterruptCapability(ctx)

	// Run error handling pattern example
	errorHandlingPattern(ctx)
}

func basicStreaming(ctx context.Context) {
	fmt.Println("=== BASIC STREAMING ===")

	client, err := claude.NewClient(ctx, types.NewClaudeAgentOptions())
	if err != nil {
		log.Printf("Failed to create client: %v", err)
		return
	}

	if err := client.Connect(ctx); err != nil {
		log.Printf("Failed to connect: %v", err)
		return
	}
	defer client.Close(ctx)

	fmt.Println("User: What is 2+2?")
	if err := client.Query(ctx, "What is 2+2?"); err != nil {
		log.Printf("Query failed: %v", err)
		return
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
}

func streamingWithRealTimeDisplay(ctx context.Context) {
	fmt.Println("=== STREAMING WITH REAL-TIME DISPLAY ===")

	client, err := claude.NewClient(ctx, types.NewClaudeAgentOptions())
	if err != nil {
		log.Printf("Failed to create client: %v", err)
		return
	}

	if err := client.Connect(ctx); err != nil {
		log.Printf("Failed to connect: %v", err)
		return
	}
	defer client.Close(ctx)

	sendAndReceive := func(prompt string) {
		fmt.Printf("User: %s\n", prompt)
		if err := client.Query(ctx, prompt); err != nil {
			log.Printf("Query failed: %v", err)
			return
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
	}

	sendAndReceive("Tell me a short joke")
	fmt.Println("\n---")
	sendAndReceive("Now tell me a fun fact")

	fmt.Println()
}

func persistentClientExample(ctx context.Context) {
	fmt.Println("=== PERSISTENT CLIENT FOR MULTIPLE QUESTIONS ===")

	// Create client
	client, err := claude.NewClient(ctx, types.NewClaudeAgentOptions())
	if err != nil {
		log.Printf("Failed to create client: %v", err)
		return
	}

	if err := client.Connect(ctx); err != nil {
		log.Printf("Failed to connect: %v", err)
		return
	}

	// Helper to get response
	getResponse := func() {
		for msg := range client.ReceiveResponse(ctx) {
			if assistantMsg, ok := msg.(*types.AssistantMessage); ok {
				for _, block := range assistantMsg.Content {
					if textBlock, ok := block.(*types.TextBlock); ok {
						fmt.Printf("Claude: %s\n", textBlock.Text)
					}
				}
			}
		}
	}

	// Use it multiple times
	fmt.Println("User: What's 2+2?")
	if err := client.Query(ctx, "What's 2+2?"); err != nil {
		log.Printf("Query failed: %v", err)
		return
	}
	getResponse()

	fmt.Println("User: What's 10*10?")
	if err := client.Query(ctx, "What's 10*10?"); err != nil {
		log.Printf("Query failed: %v", err)
		return
	}
	getResponse()

	// Don't forget to close when done
	client.Close(ctx)

	fmt.Println()
}

func withInterruptCapability(ctx context.Context) {
	fmt.Println("=== WITH INTERRUPT CAPABILITY ===")
	// IMPORTANT: Interrupts require active message consumption. You must be
	// consuming messages from the client for the interrupt to be processed.

	client, err := claude.NewClient(ctx, types.NewClaudeAgentOptions())
	if err != nil {
		log.Printf("Failed to create client: %v", err)
		return
	}

	if err := client.Connect(ctx); err != nil {
		log.Printf("Failed to connect: %v", err)
		return
	}

	fmt.Println("\n--- Sending initial message ---")

	// Send a long-running task
	fmt.Println("User: Count from 1 to 20, run bash sleep for 1 second in between")
	if err := client.Query(ctx, "Count from 1 to 20, run bash sleep for 1 second in between"); err != nil {
		log.Printf("Query failed: %v", err)
		return
	}

	// Create a background goroutine to consume messages
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

	// Wait a bit then send interrupt
	time.Sleep(10 * time.Second)
	fmt.Println("\n--- Sending interrupt ---")
	if err := client.Interrupt(ctx); err != nil {
		log.Printf("Interrupt failed: %v", err)
		return
	}

	// Wait for the background goroutine to finish
	<-done

	// Send a new message after interrupt
	fmt.Println("\n--- After interrupt, sending new message ---")
	if err := client.Query(ctx, "Just say 'Hello! I was interrupted.'"); err != nil {
		log.Printf("Query failed: %v", err)
		return
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

	client.Close(ctx)

	fmt.Println()
}

func errorHandlingPattern(ctx context.Context) {
	fmt.Println("=== ERROR HANDLING PATTERN ===")

	client, err := claude.NewClient(ctx, types.NewClaudeAgentOptions())
	if err != nil {
		log.Printf("Failed to create client: %v", err)
		return
	}

	if err := client.Connect(ctx); err != nil {
		log.Printf("Failed to connect: %v", err)
		return
	}

	// Send a message that will take time to process
	fmt.Println("User: Run a bash sleep command for 60 seconds")
	if err := client.Query(ctx, "Run a bash sleep command for 60 seconds"); err != nil {
		log.Printf("Query failed: %v", err)
		return
	}

	// Create a timeout context for receiving responses
	timeoutCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	// Receive messages with timeout
	messagesReceived := 0
	for {
		select {
		case msg, ok := <-client.ReceiveResponse(timeoutCtx):
			if !ok {
				fmt.Println("Response channel closed")
				goto done
			}
			messagesReceived++
			if assistantMsg, ok := msg.(*types.AssistantMessage); ok {
				for _, block := range assistantMsg.Content {
					if textBlock, ok := block.(*types.TextBlock); ok {
						fmt.Printf("Claude: %s\n", textBlock.Text[:min(len(textBlock.Text), 50)]+"...")
					}
				}
			}
		case <-timeoutCtx.Done():
			fmt.Println("\nRequest timed out after 20 seconds")
			fmt.Printf("Received %d messages before timeout\n", messagesReceived)
			goto done
		}
	}

done:
	client.Close(ctx)

	fmt.Println()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
