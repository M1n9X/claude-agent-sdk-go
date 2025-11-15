package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	claude "github.com/M1n9X/claude-agent-sdk-go"
	"github.com/M1n9X/claude-agent-sdk-go/types"
)

// ToolPermissionCallback demonstrates how to use tool permission callbacks to control
// which tools Claude can use and modify their inputs.
func main() {
	ctx := context.Background()

	if err := mainExample(ctx); err != nil {
		log.Printf("Tool permission callback example failed: %v", err)
	}
}

func mainExample(ctx context.Context) error {
	fmt.Println("=")
	fmt.Println("Tool Permission Callback Example")
	fmt.Println("=")
	fmt.Println("\nThis example demonstrates how to:")
	fmt.Println("1. Allow/deny tools based on type")
	fmt.Println("2. Modify tool inputs for safety")
	fmt.Println("3. Log tool usage")
	fmt.Println("4. Prompt for unknown tools")
	fmt.Println("=")

	// Configure options with our callback
	opts := types.NewClaudeAgentOptions().
		// WithModel("claude-sonnet-4-5-20250929").
		WithCanUseTool(myPermissionCallback)

	// Create client and send a query that will use multiple tools
	client, err := claude.NewClient(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	if err := client.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer client.Close(ctx)

	fmt.Println("\n📝 Sending query to Claude...")
	query := "Please do the following:\n" +
		"1. List the files in the current directory\n" +
		"2. Create a simple Python hello world script at hello.py\n" +
		"3. Run the script to test it"

	if err := client.Query(ctx, query); err != nil {
		return fmt.Errorf("query failed: %w", err)
	}

	fmt.Println("\n📨 Receiving response...")
	messageCount := 0

	for msg := range client.ReceiveResponse(ctx) {
		messageCount++

		if assistantMsg, ok := msg.(*types.AssistantMessage); ok {
			// Print Claude's text responses
			for _, block := range assistantMsg.Content {
				if textBlock, ok := block.(*types.TextBlock); ok {
					fmt.Printf("\n💬 Claude: %s\n", textBlock.Text)
				}
			}
		} else if resultMsg, ok := msg.(*types.ResultMessage); ok {
			fmt.Println("\n✅ Task completed!")
			fmt.Printf("   Duration: %dms\n", resultMsg.DurationMs)
			if resultMsg.TotalCostUSD != nil && *resultMsg.TotalCostUSD > 0 {
				fmt.Printf("   Cost: $%.4f\n", *resultMsg.TotalCostUSD)
			}
			fmt.Printf("   Messages processed: %d\n", messageCount)
		}
	}

	return nil
}

// myPermissionCallback controls tool permissions based on tool type and input.
func myPermissionCallback(ctx context.Context, toolName string, input map[string]interface{}, permCtx types.ToolPermissionContext) (interface{}, error) {
	fmt.Printf("\n🔧 Tool Permission Request: %s\n", toolName)
	fmt.Printf("   Input: %+v\n", input)

	// Always allow read operations
	if toolName == "Read" || toolName == "Glob" || toolName == "Grep" {
		fmt.Printf("   ✅ Automatically allowing %s (read-only operation)\n", toolName)
		return map[string]interface{}{
			"permissionDecision": "allow",
		}, nil
	}

	// Deny write operations to system directories
	if toolName == "Write" || toolName == "Edit" || toolName == "MultiEdit" {
		if filePath, ok := input["file_path"].(string); ok {
			if strings.HasPrefix(filePath, "/etc/") || strings.HasPrefix(filePath, "/usr/") {
				fmt.Printf("   ❌ Denying write to system directory: %s\n", filePath)
				return map[string]interface{}{
					"permissionDecision":       "deny",
					"permissionDecisionReason": fmt.Sprintf("Cannot write to system directory: %s", filePath),
				}, nil
			}

			// Redirect writes to a safe directory
			if !strings.HasPrefix(filePath, "/tmp/") && !strings.HasPrefix(filePath, "./") {
				safePath := fmt.Sprintf("./safe_output/%s", filePath[strings.LastIndex(filePath, "/")+1:])
				fmt.Printf("   ⚠️  Redirecting write from %s to %s\n", filePath, safePath)
				// Create a copy of the input with the modified path
				modifiedInput := make(map[string]interface{})
				for k, v := range input {
					modifiedInput[k] = v
				}
				modifiedInput["file_path"] = safePath
				return map[string]interface{}{
					"permissionDecision": "allow",
					"updatedInput":       modifiedInput,
				}, nil
			}
		}
	}

	// Check dangerous bash commands
	if toolName == "Bash" {
		if command, ok := input["command"].(string); ok {
			dangerousCommands := []string{"rm -rf", "sudo", "chmod 777", "dd if=", "mkfs"}

			for _, dangerous := range dangerousCommands {
				if strings.Contains(command, dangerous) {
					fmt.Printf("   ❌ Denying dangerous command: %s\n", command)
					return map[string]interface{}{
						"permissionDecision":       "deny",
						"permissionDecisionReason": fmt.Sprintf("Dangerous command pattern detected: %s", dangerous),
					}, nil
				}
			}

			// Allow but log the command
			fmt.Printf("   ✅ Allowing bash command: %s\n", command)
			return map[string]interface{}{
				"permissionDecision": "allow",
			}, nil
		}
	}

	// For all other tools, ask the user
	fmt.Printf("   ❓ Unknown tool: %s\n", toolName)
	fmt.Printf("      Input: %+v\n", input)

	fmt.Print("   Allow this tool? (y/N): ")
	reader := bufio.NewReader(os.Stdin)
	userInput, _ := reader.ReadString('\n')
	userInput = strings.TrimSpace(strings.ToLower(userInput))

	if userInput == "y" || userInput == "yes" {
		return map[string]interface{}{
			"permissionDecision": "allow",
		}, nil
	} else {
		return map[string]interface{}{
			"permissionDecision":       "deny",
			"permissionDecisionReason": "User denied permission",
		}, nil
	}
}
