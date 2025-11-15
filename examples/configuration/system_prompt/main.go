package main

import (
	"context"
	"fmt"
	"log"

	claude "github.com/M1n9X/claude-agent-sdk-go"
	"github.com/M1n9X/claude-agent-sdk-go/types"
)

// SystemPrompt demonstrates how to configure Claude with custom system prompts
// to guide its behavior and responses.
func main() {
	ctx := context.Background()

	// Example 1: Simple string system prompt
	fmt.Println("=== Example 1: String System Prompt ===")
	opts1 := types.NewClaudeAgentOptions().
		WithSystemPrompt("You are a helpful assistant that explains things simply and concisely.")

	messages, err := claude.Query(ctx, "What is Python?", opts1)
	if err != nil {
		log.Printf("Query failed: %v", err)
	} else {
		for msg := range messages {
			if msg.GetMessageType() == "assistant" {
				if assistantMsg, ok := msg.(*types.AssistantMessage); ok {
					for _, block := range assistantMsg.Content {
						if textBlock, ok := block.(*types.TextBlock); ok {
							fmt.Printf("Claude: %s\n", textBlock.Text)
						}
					}
				}
			}
		}
	}

	fmt.Println("\n" + "==================================================" + "\n")

	// Example 2: System prompt with preset
	fmt.Println("=== Example 2: System Prompt Preset ===")
	preset := types.SystemPromptPreset{
		Type:   "preset",
		Preset: "claude_code",                             // Use Claude Code's built-in preset
		Append: &[]string{"But also be very concise."}[0], // Additional custom instructions
	}

	opts2 := types.NewClaudeAgentOptions().
		WithSystemPromptPreset(preset)

	messages, err = claude.Query(ctx, "How do I create a simple HTTP server in Python?", opts2)
	if err != nil {
		log.Printf("Query failed: %v", err)
	} else {
		for msg := range messages {
			if msg.GetMessageType() == "assistant" {
				if assistantMsg, ok := msg.(*types.AssistantMessage); ok {
					for _, block := range assistantMsg.Content {
						if textBlock, ok := block.(*types.TextBlock); ok {
							fmt.Printf("Claude: %s\n", textBlock.Text)
						}
					}
				}
			}
		}
	}

	fmt.Println("\n" + "==================================================" + "\n")

	// Example 3: System prompt with tools context and fallback model
	fmt.Println("=== Example 3: System Prompt with Tools Context and Fallback Model ===")
	opts3 := types.NewClaudeAgentOptions().
		WithSystemPrompt("You are a coding assistant. You can use the Bash, Read, and Write tools to help with programming tasks. Always explain your approach before using tools.").
		// WithModel("claude-sonnet-4-5-20250929").
		WithFallbackModel("claude-3-5-haiku-latest"). // Fallback if primary model unavailable
		WithAllowedTools("Bash", "Read", "Write")

	messages, err = claude.Query(ctx, "Show me how to check the current directory and list files", opts3)
	if err != nil {
		log.Printf("Query failed: %v", err)
	} else {
		for msg := range messages {
			if msg.GetMessageType() == "assistant" {
				if assistantMsg, ok := msg.(*types.AssistantMessage); ok {
					for _, block := range assistantMsg.Content {
						if textBlock, ok := block.(*types.TextBlock); ok {
							fmt.Printf("Claude: %s\n", textBlock.Text)
						} else if toolUseBlock, ok := block.(*types.ToolUseBlock); ok {
							fmt.Printf("Tool Use: %s with input: %v\n", toolUseBlock.Name, toolUseBlock.Input)
						}
					}
				}
			}
		}
	}
}
