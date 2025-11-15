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
		return &types.PermissionResultAllow{
			Behavior: "allow",
		}, nil
	}

	// Handle write operations with user approval
	if toolName == "Write" || toolName == "Edit" || toolName == "MultiEdit" {
		if filePath, ok := input["file_path"].(string); ok {
			// Deny write operations to system directories
			if strings.HasPrefix(filePath, "/etc/") || strings.HasPrefix(filePath, "/usr/") {
				fmt.Printf("   ❌ Denying write to system directory: %s\n", filePath)
				return &types.PermissionResultDeny{
					Behavior:  "deny",
					Message:   fmt.Sprintf("Cannot write to system directory: %s", filePath),
					Interrupt: false,
				}, nil
			}

			// For other write operations, ask the user
			fmt.Printf("   ❓ Write operation requested: %s\n", filePath)
			fmt.Print("   Allow this write operation? (y/N): ")
			reader := bufio.NewReader(os.Stdin)
			userInput, _ := reader.ReadString('\n')
			userInput = strings.TrimSpace(strings.ToLower(userInput))

			if userInput == "y" || userInput == "yes" {
				// Check if we need to redirect to safe directory
				if !strings.HasPrefix(filePath, "/tmp/") && !strings.HasPrefix(filePath, "./") {
					safePath := fmt.Sprintf("./safe_output/%s", filePath[strings.LastIndex(filePath, "/")+1:])
					fmt.Printf("   ⚠️  Redirecting write from %s to %s\n", filePath, safePath)
					// Create a copy of the input with the modified path
					modifiedInput := make(map[string]interface{})
					for k, v := range input {
						modifiedInput[k] = v
					}
					modifiedInput["file_path"] = safePath
					return &types.PermissionResultAllow{
						Behavior:     "allow",
						UpdatedInput: &modifiedInput,
					}, nil
				}
				// Otherwise, allow the write as-is
				return &types.PermissionResultAllow{
					Behavior: "allow",
				}, nil
			} else {
				return &types.PermissionResultDeny{
					Behavior:  "deny",
					Message:   "User denied write permission",
					Interrupt: false,
				}, nil
			}
		} else {
			// If no file_path, ask the user
			fmt.Printf("   ❓ Write operation requested (no file path specified)\n")
			fmt.Print("   Allow this write operation? (y/N): ")
			reader := bufio.NewReader(os.Stdin)
			userInput, _ := reader.ReadString('\n')
			userInput = strings.TrimSpace(strings.ToLower(userInput))

			if userInput == "y" || userInput == "yes" {
				return &types.PermissionResultAllow{
					Behavior: "allow",
				}, nil
			} else {
				return &types.PermissionResultDeny{
					Behavior:  "deny",
					Message:   "User denied write permission",
					Interrupt: false,
				}, nil
			}
		}
	}

	// Handle bash commands with user approval
	if toolName == "Bash" {
		if command, ok := input["command"].(string); ok {
			dangerousCommands := []string{"rm -rf", "sudo", "chmod 777", "dd if=", "mkfs"}

			// Check for dangerous commands
			for _, dangerous := range dangerousCommands {
				if strings.Contains(command, dangerous) {
					fmt.Printf("   ❌ Denying dangerous command: %s\n", command)
					return &types.PermissionResultDeny{
						Behavior:  "deny",
						Message:   fmt.Sprintf("Dangerous command pattern detected: %s", dangerous),
						Interrupt: false,
					}, nil
				}
			}

			// For all bash commands (even safe ones), ask the user
			fmt.Printf("   ❓ Bash command requested: %s\n", command)
			fmt.Print("   Allow this bash command? (y/N): ")
			reader := bufio.NewReader(os.Stdin)
			userInput, _ := reader.ReadString('\n')
			userInput = strings.TrimSpace(strings.ToLower(userInput))

			if userInput == "y" || userInput == "yes" {
				fmt.Printf("   ✅ Allowing bash command: %s\n", command)
				return &types.PermissionResultAllow{
					Behavior: "allow",
				}, nil
			} else {
				return &types.PermissionResultDeny{
					Behavior:  "deny",
					Message:   "User denied bash command permission",
					Interrupt: false,
				}, nil
			}
		} else {
			// If no command specified, ask the user
			fmt.Printf("   ❓ Bash command requested (no command specified)\n")
			fmt.Print("   Allow this bash command? (y/N): ")
			reader := bufio.NewReader(os.Stdin)
			userInput, _ := reader.ReadString('\n')
			userInput = strings.TrimSpace(strings.ToLower(userInput))

			if userInput == "y" || userInput == "yes" {
				return &types.PermissionResultAllow{
					Behavior: "allow",
				}, nil
			} else {
				return &types.PermissionResultDeny{
					Behavior:  "deny",
					Message:   "User denied bash command permission",
					Interrupt: false,
				}, nil
			}
		}
	}

	// For all other tools, ask the user
	fmt.Printf("   ❓ Unknown/Other tool: %s\n", toolName)
	fmt.Printf("      Input: %+v\n", input)

	fmt.Print("   Allow this tool? (y/N): ")
	reader := bufio.NewReader(os.Stdin)
	userInput, _ := reader.ReadString('\n')
	userInput = strings.TrimSpace(strings.ToLower(userInput))

	if userInput == "y" || userInput == "yes" {
		return &types.PermissionResultAllow{
			Behavior: "allow",
		}, nil
	} else {
		return &types.PermissionResultDeny{
			Behavior:  "deny",
			Message:   "User denied permission",
			Interrupt: false,
		}, nil
	}
}
