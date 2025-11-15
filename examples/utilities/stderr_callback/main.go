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

// StderrCallback demonstrates how to capture and handle stderr output
// from the Claude Code CLI process for debugging and monitoring.
func main() {
	ctx := context.Background()

	fmt.Println("Stderr Callback Example")
	fmt.Println("=======================")
	fmt.Print("This example shows how to capture stderr output from the Claude CLI.\n\n")

	// Example 1: Basic stderr callback
	fmt.Println("=== Example 1: Basic Stderr Callback ===")

	// Create a channel to receive stderr lines
	stderrLines := make(chan string, 100)

	// Define the stderr callback function
	stderrCallback := func(line string) {
		stderrLines <- fmt.Sprintf("[STDERR] %s", strings.TrimSpace(line))
	}

	opts1 := types.NewClaudeAgentOptions().
		// WithModel("claude-sonnet-4-5-20250929").
		WithStderr(stderrCallback)

	fmt.Println("Query: Simple question with stderr monitoring...")

	// Start a goroutine to print stderr messages as they arrive
	go func() {
		for line := range stderrLines {
			fmt.Printf("Monitor: %s\n", line)
		}
	}()

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

	// Close the stderr channel after a short delay
	time.Sleep(100 * time.Millisecond)
	close(stderrLines)

	fmt.Println("\n" + "==================================================")
	fmt.Println()

	// Example 2: Advanced stderr callback with filtering
	fmt.Println("=== Example 2: Advanced Stderr Callback with Filtering ===")

	logEntries := make(chan string, 100)

	// Advanced stderr callback that filters and categorizes messages
	advancedStderrCallback := func(line string) {
		line = strings.TrimSpace(line)

		// Filter and categorize stderr messages
		if strings.Contains(line, "DEBUG") || strings.Contains(line, "debug") {
			logEntries <- fmt.Sprintf("[DEBUG] %s", line)
		} else if strings.Contains(line, "WARN") || strings.Contains(line, "warning") {
			logEntries <- fmt.Sprintf("[WARNING] %s", line)
		} else if strings.Contains(line, "ERROR") || strings.Contains(line, "error") {
			logEntries <- fmt.Sprintf("[ERROR] %s", line)
		} else if strings.Contains(line, "INFO") || strings.Contains(line, "info") {
			logEntries <- fmt.Sprintf("[INFO] %s", line)
		} else {
			logEntries <- fmt.Sprintf("[OTHER] %s", line)
		}
	}

	opts2 := types.NewClaudeAgentOptions().
		// WithModel("claude-sonnet-4-5-20250929").
		WithStderr(advancedStderrCallback).
		WithAllowedTools("Bash")

	// Start a goroutine to process categorized stderr messages
	go func() {
		for entry := range logEntries {
			fmt.Printf("Log: %s\n", entry)
		}
	}()

	fmt.Println("Query: Task that may generate CLI logs...")
	messages2, err := claude.Query(ctx, "Show the current directory and environment", opts2)
	if err != nil {
		log.Printf("Query failed: %v", err)
	} else {
		for msg := range messages2 {
			if msg.GetMessageType() == "assistant" {
				if assistantMsg, ok := msg.(*types.AssistantMessage); ok {
					for _, block := range assistantMsg.Content {
						if textBlock, ok := block.(*types.TextBlock); ok {
							fmt.Printf("Claude: %s\n", textBlock.Text)
						} else if toolUseBlock, ok := block.(*types.ToolUseBlock); ok {
							fmt.Printf("Tool Use: %s\n", toolUseBlock.Name)
						}
					}
				}
			} else if msg.GetMessageType() == "user" {
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
			}
		}
	}

	time.Sleep(100 * time.Millisecond)
	close(logEntries)

	fmt.Println("\n" + "==================================================")
	fmt.Println()

	// Example 3: Stderr callback for debugging and diagnostics
	fmt.Println("=== Example 3: Diagnostic Stderr Callback ===")

	diagnosticLog := []string{}

	diagnosticCallback := func(line string) {
		line = strings.TrimSpace(line)
		diagnosticLog = append(diagnosticLog, fmt.Sprintf("%s - %s", time.Now().Format("15:04:05"), line))
	}

	opts3 := types.NewClaudeAgentOptions().
		WithStderr(diagnosticCallback).
		WithVerbose(true) // Enable verbose logging if supported

	fmt.Println("Query: Diagnostic test...")
	messages3, err := claude.Query(ctx, "Hello", opts3)
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

	fmt.Printf("\nDiagnostic log entries (%d total):\n", len(diagnosticLog))
	for i, entry := range diagnosticLog {
		if i < 5 { // Only show first 5 entries to avoid too much output
			fmt.Printf("  %s\n", entry)
		}
	}
	if len(diagnosticLog) > 5 {
		fmt.Printf("  ... and %d more entries\n", len(diagnosticLog)-5)
	}

	fmt.Println("\n" + "==================================================")
	fmt.Println()
	fmt.Println("Stderr Callback Summary:")
	fmt.Println("- Use WithStderr() to set a callback function for stderr output")
	fmt.Println("- Useful for debugging CLI process issues")
	fmt.Println("- Can capture performance metrics, errors, and warnings")
	fmt.Println("- Callback runs in separate goroutine to avoid blocking")
	fmt.Println("- Great for monitoring and logging Claude CLI activity")
}
