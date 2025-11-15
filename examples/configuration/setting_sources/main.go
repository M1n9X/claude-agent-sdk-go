package main

import (
	"context"
	"fmt"
	"log"

	claude "github.com/M1n9X/claude-agent-sdk-go"
	"github.com/M1n9X/claude-agent-sdk-go/types"
)

// SettingSources demonstrates how to configure which setting sources Claude loads from
// to control configuration inheritance and precedence.
func main() {
	ctx := context.Background()

	fmt.Println("Setting Sources Example")
	fmt.Println("=======================")
	fmt.Print("This example shows how to control which configuration sources Claude uses.\n\n")

	// Example 1: Default setting sources
	fmt.Println("=== Example 1: Default Setting Sources ===")
	fmt.Println("By default, Claude loads settings from all available sources...")

	opts1 := types.NewClaudeAgentOptions().
		WithModel("claude-sonnet-4-5-20250929")

	fmt.Println("Query: What is the current configuration?")
	messages1, err := claude.Query(ctx, "What is 2 + 2?", opts1)
	if err != nil {
		log.Printf("Query failed: %v", err)
	} else {
		for msg := range messages1 {
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

	fmt.Println("\n" + "==================================================")
	fmt.Println()

	// Example 2: User setting source only
	fmt.Println("=== Example 2: User Setting Source Only ===")
	fmt.Println("Load settings only from user-specific configuration...")

	opts2 := types.NewClaudeAgentOptions().
		WithSettingSources(types.SettingSourceUser)

	fmt.Println("Query: Respond with your configuration context")
	messages2, err := claude.Query(ctx, "Explain what you know about your current configuration", opts2)
	if err != nil {
		log.Printf("Query failed: %v", err)
	} else {
		for msg := range messages2 {
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

	fmt.Println("\n" + "==================================================")
	fmt.Println()

	// Example 3: Project setting source only
	fmt.Println("=== Example 3: Project Setting Source Only ===")
	fmt.Println("Load settings only from project-specific configuration...")

	opts3 := types.NewClaudeAgentOptions().
		WithSettingSources(types.SettingSourceProject)

	fmt.Println("Query: Describe project-level configuration")
	messages3, err := claude.Query(ctx, "What configuration would be appropriate for a project setting?", opts3)
	if err != nil {
		log.Printf("Query failed: %v", err)
	} else {
		for msg := range messages3 {
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

	fmt.Println("\n" + "==================================================")
	fmt.Println()

	// Example 4: Multiple setting sources
	fmt.Println("=== Example 4: Multiple Setting Sources ===")
	fmt.Println("Load settings from multiple sources with specific precedence...")

	opts4 := types.NewClaudeAgentOptions().
		WithSettingSources(types.SettingSourceUser, types.SettingSourceProject)

	fmt.Println("Query: How do multiple configuration sources work?")
	messages4, err := claude.Query(ctx, "Explain how configuration settings from different sources might interact", opts4)
	if err != nil {
		log.Printf("Query failed: %v", err)
	} else {
		for msg := range messages4 {
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

	fmt.Println("\n" + "==================================================")
	fmt.Println()
	fmt.Println("Setting Sources Summary:")
	fmt.Println("- SettingSourceUser: Load from user-specific config")
	fmt.Println("- SettingSourceProject: Load from project-specific config")
	fmt.Println("- SettingSourceLocal: Load from local directory config")
	fmt.Println("- Use WithSettingSources() to specify which sources to use")
	fmt.Println("- This allows fine-grained control over configuration inheritance")
	fmt.Println("- Useful for environment-specific settings and security")
}
