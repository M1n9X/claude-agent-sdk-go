package main

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	claude "github.com/M1n9X/claude-agent-sdk-go"
	"github.com/M1n9X/claude-agent-sdk-go/types"
)

// PluginExample demonstrates how to use plugins with the Claude Go SDK.
// Plugins allow you to extend Claude Code with custom commands, agents, skills,
// and hooks. This example shows how to load a local plugin.
func main() {
	ctx := context.Background()

	fmt.Println("Plugin Example")
	fmt.Println("==============")
	fmt.Print("This example shows how to load plugins to extend Claude functionality.\n\n")

	// Example 1: Basic plugin loading
	fmt.Println("=== Example 1: Basic Plugin Loading ===")

	// Get the path to a demo plugin (relative to the current directory)
	// In production, you can use any path to your plugin directory
	pluginPath := filepath.Join("examples", "plugins", "demo-plugin")

	plugin := types.SdkPluginConfig{
		Type: "local",
		Path: pluginPath,
	}

	opts1 := types.NewClaudeAgentOptions().
		WithPlugin(plugin).
		WithMaxTurns(1) // Limit to one turn for quick demo

	fmt.Printf("Loading plugin from: %s\n", pluginPath)
	fmt.Println("Query: Hello!")

	messages1, err := claude.Query(ctx, "Hello!", opts1)
	if err != nil {
		log.Printf("Query failed: %v", err)
	} else {
		foundSystemInit := false
		for msg := range messages1 {
			msgType := msg.GetMessageType()
			if msgType == "system" {
				if systemMsg, ok := msg.(*types.SystemMessage); ok {
					if systemMsg.Subtype == "init" {
						fmt.Println("\nSystem initialized!")
						fmt.Println()
						fmt.Printf("System message data keys: %v\n", getMapKeys(systemMsg.Data))

						// Check for plugins in the system message data
						if pluginsData, ok := systemMsg.Data["plugins"]; ok {
							fmt.Println("\nPlugins loaded from system message:")
							fmt.Println()
							if pluginsSlice, ok := pluginsData.([]interface{}); ok {
								for _, plugin := range pluginsSlice {
									fmt.Printf("  - %v\n", plugin)
								}
							} else {
								fmt.Printf("  Plugin data: %v\n", pluginsData)
							}
							foundSystemInit = true
						} else {
							fmt.Println("\nNote: Plugin was passed via CLI but may not appear in system message.")
							fmt.Println()
							fmt.Printf("Plugin path configured: %s\n", pluginPath)
							foundSystemInit = true
						}
					}
				}
			} else if msgType == "assistant" {
				if assistantMsg, ok := msg.(*types.AssistantMessage); ok {
					for _, block := range assistantMsg.Content {
						if textBlock, ok := block.(*types.TextBlock); ok {
							fmt.Printf("Claude: %s\n", textBlock.Text)
						}
					}
				}
			}
		}

		if foundSystemInit {
			fmt.Println("\nPlugin successfully configured!")
			fmt.Println()
		}
	}

	fmt.Println("\n" + "==================================================")
	fmt.Println()

	// Example 2: Multiple plugins
	fmt.Println("=== Example 2: Multiple Plugins ===")

	// Note: In a real scenario, you'd have multiple plugin directories
	// For this example, we'll use the same plugin twice to demonstrate the concept
	plugin1 := types.SdkPluginConfig{
		Type: "local",
		Path: pluginPath,
	}

	plugin2 := types.SdkPluginConfig{
		Type: "local",
		Path: pluginPath, // Same path for demo purposes
	}

	opts2 := types.NewClaudeAgentOptions().
		WithPlugins([]types.SdkPluginConfig{plugin1, plugin2}).
		WithMaxTurns(1)

	fmt.Println("Loading multiple plugins...")
	fmt.Println("Query: Hello again!")

	messages2, err := claude.Query(ctx, "Hello again!", opts2)
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
	fmt.Println("Plugin Support Summary:")
	fmt.Println("- Use WithPlugin() or WithPlugins() to configure plugins")
	fmt.Println("- Currently supports 'local' type plugins")
	fmt.Println("- Plugins extend Claude with custom commands and functionality")
	fmt.Println("- Plugin paths should point to plugin directories")
	fmt.Println("- Plugins are loaded before Claude session starts")
}

// Helper function to get map keys for display
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
