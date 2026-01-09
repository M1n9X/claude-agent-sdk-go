package main

import (
	"context"
	"fmt"
	"log"
	"time"

	claude "github.com/M1n9X/claude-agent-sdk-go"
	"github.com/M1n9X/claude-agent-sdk-go/types"
)

// ComprehensiveHooks demonstrates all available hook events in the SDK.
// This example shows how to use hooks for monitoring, logging, and controlling
// the agent's behavior at various lifecycle points.
func main() {
	ctx := context.Background()

	// Create hook matchers for all hook events
	opts := types.NewClaudeAgentOptions().
		WithAllowedTools("Bash", "Read", "Write").
		WithPermissionMode(types.PermissionModeBypassPermissions).
		// PreToolUse: Called before a tool is executed
		WithHook(types.HookEventPreToolUse, types.HookMatcher{
			Hooks: []types.HookCallbackFunc{preToolUseHook},
		}).
		// PostToolUse: Called after a tool execution completes
		WithHook(types.HookEventPostToolUse, types.HookMatcher{
			Hooks: []types.HookCallbackFunc{postToolUseHook},
		}).
		// UserPromptSubmit: Called when user submits a prompt
		WithHook(types.HookEventUserPromptSubmit, types.HookMatcher{
			Hooks: []types.HookCallbackFunc{userPromptSubmitHook},
		}).
		// PrePrompt: Called before sending messages to the model
		WithHook(types.HookEventPrePrompt, types.HookMatcher{
			Hooks: []types.HookCallbackFunc{prePromptHook},
		}).
		// PostPrompt: Called after receiving response from the model
		WithHook(types.HookEventPostPrompt, types.HookMatcher{
			Hooks: []types.HookCallbackFunc{postPromptHook},
		}).
		// PreResponse: Called before sending response to user
		WithHook(types.HookEventPreResponse, types.HookMatcher{
			Hooks: []types.HookCallbackFunc{preResponseHook},
		}).
		// PostResponse: Called after sending response to user
		WithHook(types.HookEventPostResponse, types.HookMatcher{
			Hooks: []types.HookCallbackFunc{postResponseHook},
		}).
		// PreCompact: Called before context compaction
		WithHook(types.HookEventPreCompact, types.HookMatcher{
			Hooks: []types.HookCallbackFunc{preCompactHook},
		}).
		// PostCompact: Called after context compaction
		WithHook(types.HookEventPostCompact, types.HookMatcher{
			Hooks: []types.HookCallbackFunc{postCompactHook},
		}).
		// OnError: Called when an error occurs
		WithHook(types.HookEventOnError, types.HookMatcher{
			Hooks: []types.HookCallbackFunc{onErrorHook},
		}).
		// Stop: Called when the agent stops
		WithHook(types.HookEventStop, types.HookMatcher{
			Hooks: []types.HookCallbackFunc{stopHook},
		})

	fmt.Println("Comprehensive Hooks Example")
	fmt.Println("===========================")
	fmt.Println("\nThis example demonstrates all available hook events:")
	fmt.Println("1. PreToolUse - Before tool execution")
	fmt.Println("2. PostToolUse - After tool execution")
	fmt.Println("3. UserPromptSubmit - When user submits a prompt")
	fmt.Println("4. PrePrompt - Before sending to model")
	fmt.Println("5. PostPrompt - After receiving from model")
	fmt.Println("6. PreResponse - Before sending response to user")
	fmt.Println("7. PostResponse - After sending response to user")
	fmt.Println("8. PreCompact - Before context compaction")
	fmt.Println("9. PostCompact - After context compaction")
	fmt.Println("10. OnError - When an error occurs")
	fmt.Println("11. Stop - When the agent stops")
	fmt.Println()

	// Send query
	fmt.Println("Query: 'What is the current directory?'")
	fmt.Println("------------------------------------------------------------")

	messages, err := claude.Query(ctx, "What is the current directory?", opts)
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}

	// Process messages
	for msg := range messages {
		switch m := msg.(type) {
		case *types.AssistantMessage:
			for _, block := range m.Content {
				if tb, ok := block.(*types.TextBlock); ok {
					fmt.Printf("\n💬 Claude: %s\n", tb.Text)
				}
			}
		case *types.ResultMessage:
			fmt.Println("\n------------------------------------------------------------")
			fmt.Println("✅ Query completed")
			if m.TotalCostUSD != nil && *m.TotalCostUSD > 0 {
				fmt.Printf("💰 Cost: $%.4f\n", *m.TotalCostUSD)
			}
			fmt.Printf("⏱️  Duration: %dms\n", m.DurationMs)
			fmt.Printf("🔄 Turns: %d\n", m.NumTurns)
		}
	}

	fmt.Println("\n============================================================")
	fmt.Println("Example completed!")
}

// preToolUseHook is called before Claude uses a tool
func preToolUseHook(ctx context.Context, input interface{}, toolUseID *string, hookCtx types.HookContext) (interface{}, error) {
	fmt.Println("\n🔧 [PreToolUse Hook] Tool about to be executed")

	// Type assert to get the actual input structure
	if inputMap, ok := input.(map[string]interface{}); ok {
		if toolName, ok := inputMap["tool_name"].(string); ok {
			fmt.Printf("   Tool: %s\n", toolName)
		}
		if toolInput, ok := inputMap["tool_input"].(map[string]interface{}); ok {
			fmt.Printf("   Input: %+v\n", toolInput)
		}
	}

	// You can modify the tool input or deny execution here
	// For example, to deny execution:
	// return map[string]interface{}{
	//     "hookSpecificOutput": map[string]interface{}{
	//         "hookEventName": "PreToolUse",
	//         "permissionDecision": "deny",
	//         "permissionDecisionReason": "Tool blocked by hook",
	//     },
	// }, nil

	// Allow execution
	return &types.SyncHookJSONOutput{}, nil
}

// postToolUseHook is called after a tool execution completes
func postToolUseHook(ctx context.Context, input interface{}, toolUseID *string, hookCtx types.HookContext) (interface{}, error) {
	fmt.Println("\n✅ [PostToolUse Hook] Tool execution completed")

	if inputMap, ok := input.(map[string]interface{}); ok {
		if toolName, ok := inputMap["tool_name"].(string); ok {
			fmt.Printf("   Tool: %s\n", toolName)
		}
		if toolResponse, ok := inputMap["tool_response"]; ok {
			fmt.Printf("   Response: %+v\n", toolResponse)
		}
	}

	// You can add additional context or modify the response here
	return &types.SyncHookJSONOutput{}, nil
}

// userPromptSubmitHook is called when user submits a prompt
func userPromptSubmitHook(ctx context.Context, input interface{}, toolUseID *string, hookCtx types.HookContext) (interface{}, error) {
	fmt.Println("\n📝 [UserPromptSubmit Hook] User submitted a prompt")

	if inputMap, ok := input.(map[string]interface{}); ok {
		if prompt, ok := inputMap["prompt"].(string); ok {
			fmt.Printf("   Prompt: %s\n", prompt)
		}
	}

	return &types.SyncHookJSONOutput{}, nil
}

// prePromptHook is called before sending messages to the model
func prePromptHook(ctx context.Context, input interface{}, toolUseID *string, hookCtx types.HookContext) (interface{}, error) {
	fmt.Println("\n📤 [PrePrompt Hook] About to send messages to model")

	if inputMap, ok := input.(map[string]interface{}); ok {
		if messages, ok := inputMap["messages"].([]interface{}); ok {
			fmt.Printf("   Message count: %d\n", len(messages))
		}
	}

	// You can modify messages or add system prompts here
	return &types.SyncHookJSONOutput{}, nil
}

// postPromptHook is called after receiving response from the model
func postPromptHook(ctx context.Context, input interface{}, toolUseID *string, hookCtx types.HookContext) (interface{}, error) {
	fmt.Println("\n📥 [PostPrompt Hook] Received response from model")

	if inputMap, ok := input.(map[string]interface{}); ok {
		if response, ok := inputMap["response"].(map[string]interface{}); ok {
			if model, ok := response["model"].(string); ok {
				fmt.Printf("   Model: %s\n", model)
			}
		}
	}

	return &types.SyncHookJSONOutput{}, nil
}

// preResponseHook is called before sending response to user
func preResponseHook(ctx context.Context, input interface{}, toolUseID *string, hookCtx types.HookContext) (interface{}, error) {
	fmt.Println("\n📨 [PreResponse Hook] About to send response to user")

	// You can modify the response before it's sent to the user
	return &types.SyncHookJSONOutput{}, nil
}

// postResponseHook is called after sending response to user
func postResponseHook(ctx context.Context, input interface{}, toolUseID *string, hookCtx types.HookContext) (interface{}, error) {
	fmt.Println("\n✉️  [PostResponse Hook] Response sent to user")

	return &types.SyncHookJSONOutput{}, nil
}

// preCompactHook is called before context compaction
func preCompactHook(ctx context.Context, input interface{}, toolUseID *string, hookCtx types.HookContext) (interface{}, error) {
	fmt.Println("\n🗜️  [PreCompact Hook] About to compact context")

	if inputMap, ok := input.(map[string]interface{}); ok {
		if trigger, ok := inputMap["trigger"].(string); ok {
			fmt.Printf("   Trigger: %s\n", trigger)
		}
	}

	// You can add custom instructions for compaction here
	return &types.SyncHookJSONOutput{}, nil
}

// postCompactHook is called after context compaction
func postCompactHook(ctx context.Context, input interface{}, toolUseID *string, hookCtx types.HookContext) (interface{}, error) {
	fmt.Println("\n✅ [PostCompact Hook] Context compaction completed")

	if inputMap, ok := input.(map[string]interface{}); ok {
		if compactedTokens, ok := inputMap["compacted_tokens"].(float64); ok {
			if originalTokens, ok := inputMap["original_tokens"].(float64); ok {
				ratio := compactedTokens / originalTokens * 100
				fmt.Printf("   Compacted: %.0f tokens → %.0f tokens (%.1f%%)\n",
					originalTokens, compactedTokens, ratio)
			}
		}
	}

	return &types.SyncHookJSONOutput{}, nil
}

// onErrorHook is called when an error occurs
func onErrorHook(ctx context.Context, input interface{}, toolUseID *string, hookCtx types.HookContext) (interface{}, error) {
	fmt.Println("\n❌ [OnError Hook] An error occurred")

	if inputMap, ok := input.(map[string]interface{}); ok {
		if errorMsg, ok := inputMap["error"].(string); ok {
			fmt.Printf("   Error: %s\n", errorMsg)
		}
		if errorType, ok := inputMap["error_type"].(string); ok {
			fmt.Printf("   Type: %s\n", errorType)
		}
	}

	// You can implement error recovery logic here
	// For example, to retry:
	// return map[string]interface{}{
	//     "hookSpecificOutput": map[string]interface{}{
	//         "hookEventName": "OnError",
	//         "recoveryAction": "retry",
	//     },
	// }, nil

	return &types.SyncHookJSONOutput{}, nil
}

// stopHook is called when the agent stops
func stopHook(ctx context.Context, input interface{}, toolUseID *string, hookCtx types.HookContext) (interface{}, error) {
	fmt.Println("\n🛑 [Stop Hook] Agent is stopping")

	// You can perform cleanup or logging here
	timestamp := time.Now().Format(time.RFC3339)
	fmt.Printf("   Stopped at: %s\n", timestamp)

	return &types.SyncHookJSONOutput{}, nil
}
