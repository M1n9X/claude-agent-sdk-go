package main

import (
	"context"
	"fmt"
	"log"

	claude "github.com/M1n9X/claude-agent-sdk-go"
	"github.com/M1n9X/claude-agent-sdk-go/types"
)

// PythonEquivalence demonstrates how Go SDK APIs map to Python SDK APIs.
// This example shows side-by-side comparisons of common patterns.
func main() {
	ctx := context.Background()

	fmt.Println("Python SDK Equivalence Examples")
	fmt.Println("================================")
	fmt.Println()

	// Example 1: Basic Query
	example1BasicQuery(ctx)

	// Example 2: Interactive Client
	example2InteractiveClient(ctx)

	// Example 3: Custom Tools
	example3CustomTools(ctx)

	// Example 4: Hooks
	example4Hooks(ctx)

	// Example 5: Permissions
	example5Permissions(ctx)

	fmt.Println("\n" + "============================================================")
	fmt.Println("All examples completed!")
}

// Example 1: Basic Query
// Python:
//
//	from claude_agent_sdk import query, ClaudeAgentOptions
//	options = ClaudeAgentOptions(model="claude-sonnet-4-5")
//	async for msg in query("Hello", options):
//	    if msg.type == "assistant":
//	        print(msg.content[0].text)
func example1BasicQuery(ctx context.Context) {
	fmt.Println("Example 1: Basic Query")
	fmt.Println("----------------------")
	fmt.Println("Python equivalent:")
	fmt.Println(`  options = ClaudeAgentOptions(model="claude-sonnet-4-5")`)
	fmt.Println(`  async for msg in query("Hello", options):`)
	fmt.Println(`      if msg.type == "assistant":`)
	fmt.Println(`          print(msg.content[0].text)`)
	fmt.Println()
	fmt.Println("Go implementation:")

	opts := types.NewClaudeAgentOptions().
		WithModel("claude-sonnet-4-5-20250929")

	messages, err := claude.Query(ctx, "Say hello in one sentence", opts)
	if err != nil {
		log.Printf("Query failed: %v", err)
		return
	}

	for msg := range messages {
		if assistantMsg, ok := msg.(*types.AssistantMessage); ok {
			for _, block := range assistantMsg.Content {
				if textBlock, ok := block.(*types.TextBlock); ok {
					fmt.Printf("  Output: %s\n", textBlock.Text)
				}
			}
		}
	}
	fmt.Println()
}

// Example 2: Interactive Client
// Python:
//
//	async with ClaudeSDKClient(options=options) as client:
//	    await client.query("First question")
//	    async for msg in client.receive_response():
//	        print(msg)
func example2InteractiveClient(ctx context.Context) {
	fmt.Println("Example 2: Interactive Client")
	fmt.Println("------------------------------")
	fmt.Println("Python equivalent:")
	fmt.Println(`  async with ClaudeSDKClient(options=options) as client:`)
	fmt.Println(`      await client.query("First question")`)
	fmt.Println(`      async for msg in client.receive_response():`)
	fmt.Println(`          print(msg)`)
	fmt.Println()
	fmt.Println("Go implementation:")

	opts := types.NewClaudeAgentOptions().
		WithModel("claude-sonnet-4-5-20250929")

	client, err := claude.NewClient(ctx, opts)
	if err != nil {
		log.Printf("Failed to create client: %v", err)
		return
	}
	defer client.Close(ctx)

	if err := client.Connect(ctx); err != nil {
		log.Printf("Failed to connect: %v", err)
		return
	}

	if err := client.Query(ctx, "What is 2+2? Answer in one sentence."); err != nil {
		log.Printf("Query failed: %v", err)
		return
	}

	for msg := range client.ReceiveResponse(ctx) {
		if assistantMsg, ok := msg.(*types.AssistantMessage); ok {
			for _, block := range assistantMsg.Content {
				if textBlock, ok := block.(*types.TextBlock); ok {
					fmt.Printf("  Output: %s\n", textBlock.Text)
				}
			}
		}
	}
	fmt.Println()
}

// Example 3: Custom Tools
// Python:
//
//	@tool("greet", "Greet a user", {"name": str})
//	async def greet_user(args):
//	    return {"content": [{"type": "text", "text": f"Hello, {args['name']}!"}]}
//
//	server = create_sdk_mcp_server("tools", "1.0.0", [greet_user])
func example3CustomTools(ctx context.Context) {
	fmt.Println("Example 3: Custom Tools")
	fmt.Println("------------------------")
	fmt.Println("Python equivalent:")
	fmt.Println(`  @tool("greet", "Greet a user", {"name": str})`)
	fmt.Println(`  async def greet_user(args):`)
	fmt.Println(`      return {"content": [{"type": "text", "text": f"Hello, {args['name']}!"}]}`)
	fmt.Println()
	fmt.Println("Go implementation (Method 1 - SimpleTool):")

	// Method 1: SimpleTool (most similar to Python @tool)
	greetTool := types.SimpleTool{
		Name:        "greet",
		Description: "Greet a user",
		Parameters: map[string]types.SimpleParam{
			"name": {Type: "string", Description: "User's name", Required: true},
		},
		Handler: func(ctx context.Context, args map[string]interface{}) (*types.ToolResult, error) {
			name := args["name"].(string)
			return types.NewMcpToolResult(
				types.TextBlock{Type: "text", Text: fmt.Sprintf("Hello, %s!", name)},
			), nil
		},
	}

	tool, err := greetTool.Build()
	if err != nil {
		log.Printf("Failed to build tool: %v", err)
		return
	}

	server := types.CreateToolServer("tools", "1.0.0", []types.McpTool{tool})

	opts := types.NewClaudeAgentOptions().
		WithMcpServers(map[string]interface{}{"tools": server}).
		WithAllowedTools("mcp__tools__greet")

	messages, err := claude.Query(ctx, "Greet Alice", opts)
	if err != nil {
		log.Printf("Query failed: %v", err)
		return
	}

	for msg := range messages {
		if assistantMsg, ok := msg.(*types.AssistantMessage); ok {
			for _, block := range assistantMsg.Content {
				if textBlock, ok := block.(*types.TextBlock); ok {
					fmt.Printf("  Output: %s\n", textBlock.Text)
				}
			}
		}
	}

	fmt.Println()
	fmt.Println("Go implementation (Method 2 - QuickTool):")
	fmt.Println(`  tool, _ := types.QuickTool("greet", "Greet a user",`)
	fmt.Println(`      map[string]string{"name": "string"},`)
	fmt.Println(`      func(ctx context.Context, args map[string]interface{}) (*types.ToolResult, error) {`)
	fmt.Println(`          name := args["name"].(string)`)
	fmt.Println(`          return types.NewMcpToolResult(...)`)
	fmt.Println(`      })`)
	fmt.Println()
}

// Example 4: Hooks
// Python:
//
//	async def pre_tool_hook(input_data, tool_use_id, context):
//	    print(f"Tool {input_data['tool_name']} about to execute")
//	    return {}
//
//	options = ClaudeAgentOptions(
//	    hooks={"PreToolUse": [HookMatcher(hooks=[pre_tool_hook])]}
//	)
func example4Hooks(ctx context.Context) {
	fmt.Println("Example 4: Hooks")
	fmt.Println("----------------")
	fmt.Println("Python equivalent:")
	fmt.Println(`  async def pre_tool_hook(input_data, tool_use_id, context):`)
	fmt.Println(`      print(f"Tool {input_data['tool_name']} about to execute")`)
	fmt.Println(`      return {}`)
	fmt.Println()
	fmt.Println(`  options = ClaudeAgentOptions(`)
	fmt.Println(`      hooks={"PreToolUse": [HookMatcher(hooks=[pre_tool_hook])]}`)
	fmt.Println(`  )`)
	fmt.Println()
	fmt.Println("Go implementation:")

	preToolHook := func(ctx context.Context, input interface{}, toolUseID *string, hookCtx types.HookContext) (interface{}, error) {
		if inputMap, ok := input.(map[string]interface{}); ok {
			if toolName, ok := inputMap["tool_name"].(string); ok {
				fmt.Printf("  [Hook] Tool %s about to execute\n", toolName)
			}
		}
		return &types.SyncHookJSONOutput{}, nil
	}

	opts := types.NewClaudeAgentOptions().
		WithAllowedTools("Bash").
		WithPermissionMode(types.PermissionModeBypassPermissions).
		WithHook(types.HookEventPreToolUse, types.HookMatcher{
			Hooks: []types.HookCallbackFunc{preToolHook},
		})

	messages, err := claude.Query(ctx, "What is the current directory?", opts)
	if err != nil {
		log.Printf("Query failed: %v", err)
		return
	}

	for msg := range messages {
		if assistantMsg, ok := msg.(*types.AssistantMessage); ok {
			for _, block := range assistantMsg.Content {
				if textBlock, ok := block.(*types.TextBlock); ok {
					fmt.Printf("  Output: %s\n", textBlock.Text)
				}
			}
		}
	}
	fmt.Println()
}

// Example 5: Permissions
// Python:
//
//	async def can_use_tool(tool_name, input, context):
//	    if tool_name == "Bash":
//	        return {"behavior": "allow"}
//	    return {"behavior": "deny", "message": "Tool not allowed"}
//
//	options = ClaudeAgentOptions(can_use_tool=can_use_tool)
func example5Permissions(ctx context.Context) {
	fmt.Println("Example 5: Permissions")
	fmt.Println("----------------------")
	fmt.Println("Python equivalent:")
	fmt.Println(`  async def can_use_tool(tool_name, input, context):`)
	fmt.Println(`      if tool_name == "Bash":`)
	fmt.Println(`          return {"behavior": "allow"}`)
	fmt.Println(`      return {"behavior": "deny", "message": "Tool not allowed"}`)
	fmt.Println()
	fmt.Println(`  options = ClaudeAgentOptions(can_use_tool=can_use_tool)`)
	fmt.Println()
	fmt.Println("Go implementation:")

	canUseTool := func(ctx context.Context, toolName string, input map[string]interface{}, permCtx types.ToolPermissionContext) (interface{}, error) {
		fmt.Printf("  [Permission] Checking tool: %s\n", toolName)
		if toolName == "Bash" {
			return &types.PermissionResultAllow{Behavior: "allow"}, nil
		}
		return &types.PermissionResultDeny{
			Behavior: "deny",
			Message:  "Tool not allowed",
		}, nil
	}

	opts := types.NewClaudeAgentOptions().
		WithCanUseTool(canUseTool)

	client, err := claude.NewClient(ctx, opts)
	if err != nil {
		log.Printf("Failed to create client: %v", err)
		return
	}
	defer client.Close(ctx)

	if err := client.Connect(ctx); err != nil {
		log.Printf("Failed to connect: %v", err)
		return
	}

	if err := client.Query(ctx, "What is the current directory?"); err != nil {
		log.Printf("Query failed: %v", err)
		return
	}

	for msg := range client.ReceiveResponse(ctx) {
		if assistantMsg, ok := msg.(*types.AssistantMessage); ok {
			for _, block := range assistantMsg.Content {
				if textBlock, ok := block.(*types.TextBlock); ok {
					fmt.Printf("  Output: %s\n", textBlock.Text)
				}
			}
		}
	}
	fmt.Println()
}
