package main

import (
	"context"
	"fmt"
	"log"

	claude "github.com/M1n9X/claude-agent-sdk-go"
	"github.com/M1n9X/claude-agent-sdk-go/types"
)

// Agents demonstrates using custom agents with specific tools, prompts, and models.
func main() {
	ctx := context.Background()

	// Run code reviewer agent example
	if err := codeReviewerExample(ctx); err != nil {
		log.Printf("Code reviewer example failed: %v", err)
	}

	// Run documentation writer agent example
	if err := documentationWriterExample(ctx); err != nil {
		log.Printf("Documentation writer example failed: %v", err)
	}

	// Run multiple agents example
	if err := multipleAgentsExample(ctx); err != nil {
		log.Printf("Multiple agents example failed: %v", err)
	}
}

func codeReviewerExample(ctx context.Context) error {
	fmt.Println("=== Code Reviewer Agent Example ===")

	// Create an agent definition for code review
	model := "sonnet"
	agentDef := types.AgentDefinition{
		Description: "Reviews code for best practices and potential issues",
		Prompt:      "You are a code reviewer. Analyze code for bugs, performance issues, security vulnerabilities, and adherence to best practices. Provide constructive feedback.",
		Tools:       []string{"Read", "Grep"},
		Model:       &model,
	}

	// Create options with the agent
	opts := types.NewClaudeAgentOptions().
		// WithModel("claude-sonnet-4-5-20250929").
		WithAgent("code-reviewer", agentDef)

	// Query using the agent
	fmt.Println("User: Use the code-reviewer agent to review the code in this project")
	fmt.Println("---")

	messages, err := claude.Query(ctx, "Use the code-reviewer agent to review the code in this project", opts)
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}

	// Process messages from the channel
	for msg := range messages {
		msgType := msg.GetMessageType()

		switch msgType {
		case "assistant":
			if assistantMsg, ok := msg.(*types.AssistantMessage); ok {
				for _, block := range assistantMsg.Content {
					if textBlock, ok := block.(*types.TextBlock); ok {
						fmt.Printf("Claude: %s\n", textBlock.Text)
					}
				}
			}
		case "result":
			if resultMsg, ok := msg.(*types.ResultMessage); ok && resultMsg.TotalCostUSD != nil && *resultMsg.TotalCostUSD > 0 {
				fmt.Printf("\nCost: $%.4f\n", *resultMsg.TotalCostUSD)
			}
			fmt.Println("---")
			fmt.Println("Query completed successfully")
		}
	}

	fmt.Println()
	return nil
}

func documentationWriterExample(ctx context.Context) error {
	fmt.Println("=== Documentation Writer Agent Example ===")

	// Create an agent definition for documentation writing
	model := "sonnet"
	agentDef := types.AgentDefinition{
		Description: "Writes comprehensive documentation",
		Prompt:      "You are a technical documentation expert. Write clear, comprehensive documentation with examples. Focus on clarity and completeness.",
		Tools:       []string{"Read", "Write", "Edit"},
		Model:       &model,
	}

	// Create options with the agent
	opts := types.NewClaudeAgentOptions().
		// WithModel("claude-sonnet-4-5-20250929").
		WithAgent("doc-writer", agentDef)

	// Query using the agent
	fmt.Println("User: Use the doc-writer agent to explain what AgentDefinition is used for")
	fmt.Println("---")

	messages, err := claude.Query(ctx, "Use the doc-writer agent to explain what AgentDefinition is used for", opts)
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}

	// Process messages from the channel
	for msg := range messages {
		msgType := msg.GetMessageType()

		switch msgType {
		case "assistant":
			if assistantMsg, ok := msg.(*types.AssistantMessage); ok {
				for _, block := range assistantMsg.Content {
					if textBlock, ok := block.(*types.TextBlock); ok {
						fmt.Printf("Claude: %s\n", textBlock.Text)
					}
				}
			}
		case "result":
			if resultMsg, ok := msg.(*types.ResultMessage); ok && resultMsg.TotalCostUSD != nil && *resultMsg.TotalCostUSD > 0 {
				fmt.Printf("\nCost: $%.4f\n", *resultMsg.TotalCostUSD)
			}
			fmt.Println("---")
			fmt.Println("Query completed successfully")
		}
	}

	fmt.Println()
	return nil
}

func multipleAgentsExample(ctx context.Context) error {
	fmt.Println("=== Multiple Agents Example ===")

	// Create agent definitions for analyzer and tester
	model := "sonnet"
	analyzerDef := types.AgentDefinition{
		Description: "Analyzes code structure and patterns",
		Prompt:      "You are a code analyzer. Examine code structure, patterns, and architecture.",
		Tools:       []string{"Read", "Grep", "Glob"},
	}

	testerDef := types.AgentDefinition{
		Description: "Creates and runs tests",
		Prompt:      "You are a testing expert. Write comprehensive tests and ensure code quality.",
		Tools:       []string{"Read", "Write", "Bash"},
		Model:       &model,
	}

	// Create options with multiple agents
	opts := types.NewClaudeAgentOptions().
		// WithModel("claude-sonnet-4-5-20250929").
		WithAgent("analyzer", analyzerDef).
		WithAgent("tester", testerDef).
		WithSettingSources("user", "project")

	// Query using the agents
	fmt.Println("User: Use the analyzer agent to find all Go files in the examples/ directory")
	fmt.Println("---")

	messages, err := claude.Query(ctx, "Use the analyzer agent to find all Go files in the examples/ directory", opts)
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}

	// Process messages from the channel
	for msg := range messages {
		msgType := msg.GetMessageType()

		switch msgType {
		case "assistant":
			if assistantMsg, ok := msg.(*types.AssistantMessage); ok {
				for _, block := range assistantMsg.Content {
					if textBlock, ok := block.(*types.TextBlock); ok {
						fmt.Printf("Claude: %s\n", textBlock.Text)
					}
				}
			}
		case "result":
			if resultMsg, ok := msg.(*types.ResultMessage); ok && resultMsg.TotalCostUSD != nil && *resultMsg.TotalCostUSD > 0 {
				fmt.Printf("\nCost: $%.4f\n", *resultMsg.TotalCostUSD)
			}
			fmt.Println("---")
			fmt.Println("Query completed successfully")
		}
	}

	fmt.Println()
	return nil
}
