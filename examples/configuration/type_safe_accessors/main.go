package main

import (
	"fmt"

	"github.com/M1n9X/claude-agent-sdk-go/types"
)

// TypeSafeAccessors demonstrates how to use the new type-safe message accessors
// for cleaner and more reliable message handling.
func main() {
	fmt.Println("Type-Safe Accessors Example")
	fmt.Println("===========================")
	fmt.Print("This example shows how to use type-safe message accessors.\n\n")

	// Create some example messages
	userMsg := &types.UserMessage{
		Type:    "user",
		Content: "Hello, Claude!",
	}

	assistantMsg := &types.AssistantMessage{
		Type: "assistant",
		Content: []types.ContentBlock{
			&types.TextBlock{Type: "text", Text: "Hello! How can I help you today?"},
		},
	}

	systemMsg := &types.SystemMessage{
		Type:    "system",
		Subtype: "init",
		Data:    map[string]interface{}{"version": "1.0"},
	}

	// Example 1: Using the new AsUser function
	fmt.Println("=== Example 1: Using AsUser function ===")
	if user, ok := types.AsUser(userMsg); ok {
		fmt.Printf("Successfully extracted user message: %s\n", user.Content)
	} else {
		fmt.Println("Not a user message")
	}

	// Example 2: Using the new AsAssistant function
	fmt.Println("\n=== Example 2: Using AsAssistant function ===")
	if assistant, ok := types.AsAssistant(assistantMsg); ok {
		fmt.Printf("Successfully extracted assistant message with %d content blocks\n", len(assistant.Content))
		for _, block := range assistant.Content {
			if textBlock, ok := block.(*types.TextBlock); ok {
				fmt.Printf("  Text: %s\n", textBlock.Text)
			}
		}
	} else {
		fmt.Println("Not an assistant message")
	}

	// Example 3: Using the new AsSystem function
	fmt.Println("\n=== Example 3: Using AsSystem function ===")
	if system, ok := types.AsSystem(systemMsg); ok {
		fmt.Printf("Successfully extracted system message with subtype: %s\n", system.Subtype)
		fmt.Printf("  Data: %+v\n", system.Data)
	} else {
		fmt.Println("Not a system message")
	}

	// Example 4: Using type switch with new accessor methods
	fmt.Println("\n=== Example 4: Using accessor methods ===")
	messages := []types.Message{userMsg, assistantMsg, systemMsg}

	for i, msg := range messages {
		fmt.Printf("Message %d: ", i+1)

		// Using the new accessor methods
		if user, ok := types.AsUser(msg); ok {
			fmt.Printf("User - %v\n", user.Content)
		} else if assistant, ok := types.AsAssistant(msg); ok {
			fmt.Printf("Assistant - %d content blocks\n", len(assistant.Content))
		} else if system, ok := types.AsSystem(msg); ok {
			fmt.Printf("System - subtype: %s\n", system.Subtype)
		} else if result, ok := types.AsResult(msg); ok {
			fmt.Printf("Result - duration: %dms\n", result.DurationMs)
		} else {
			fmt.Printf("Unknown message type: %s\n", msg.GetMessageType())
		}
	}

	fmt.Println("\n" + "==================================================")
	fmt.Println()
	fmt.Println("Type-Safe Accessors Support Summary:")
	fmt.Println("- AsUser(), AsAssistant(), AsSystem(), AsResult(), etc. functions")
	fmt.Println("- Type-safe conversion without manual type assertion")
	fmt.Println("- Cleaner, more readable code")
	fmt.Println("- Compile-time safety against incorrect type assertions")
	fmt.Println("- Consistent error handling with (value, ok) pattern")
}
