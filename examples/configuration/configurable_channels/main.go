package main

import (
	"context"
	"fmt"
	"log"

	claude "github.com/M1n9X/claude-agent-sdk-go"
	"github.com/M1n9X/claude-agent-sdk-go/types"
)

// ConfigurableChannels demonstrates how to use the configurable message channel capacity
// to optimize performance based on your application's needs.
func main() {
	ctx := context.Background()

	fmt.Println("Configurable Channels Example")
	fmt.Println("=============================")
	fmt.Print("This example shows how to configure message channel capacity.\n\n")

	// Example 1: Default capacity (10 for transport, 100 for query internal)
	fmt.Println("=== Example 1: Default Capacity ===")
	fmt.Println("Using default message channel capacity...")

	opts1 := types.NewClaudeAgentOptions().
		WithModel("claude-sonnet-4-5-20250929")

	client1, err := claude.NewClient(ctx, opts1)
	if err != nil {
		log.Printf("Client creation failed (expected since CLI may not be available): %v", err)
	} else {
		fmt.Printf("Client created with default capacity\n")
		_ = client1.Close(ctx)
	}

	// Example 2: Custom capacity for high-throughput scenarios
	fmt.Println("\n=== Example 2: High-Throughput Capacity ===")
	fmt.Println("Using larger message channel capacity for high-throughput scenarios...")

	opts2 := types.NewClaudeAgentOptions().
		// WithModel("claude-sonnet-4-5-20250929").
		WithMessageChannelCapacity(500) // Larger capacity for bursty message handling

	client2, err := claude.NewClient(ctx, opts2)
	if err != nil {
		log.Printf("Client creation failed (expected since CLI may not be available): %v", err)
	} else {
		fmt.Printf("Client created with capacity 500\n")
		_ = client2.Close(ctx)
	}

	// Example 3: Custom capacity for memory-constrained scenarios
	fmt.Println("\n=== Example 3: Memory-Constrained Capacity ===")
	fmt.Println("Using smaller message channel capacity to conserve memory...")

	opts3 := types.NewClaudeAgentOptions().
		// WithModel("claude-sonnet-4-5-20250929").
		WithMessageChannelCapacity(5) // Smaller capacity to conserve memory

	client3, err := claude.NewClient(ctx, opts3)
	if err != nil {
		log.Printf("Client creation failed (expected since CLI may not be available): %v", err)
	} else {
		fmt.Printf("Client created with capacity 5\n")
		_ = client3.Close(ctx)
	}

	fmt.Println("\n" + "==================================================")
	fmt.Println()
	fmt.Println("Configurable Channels Support Summary:")
	fmt.Println("- MessageChannelCapacity option allows customizing channel buffer sizes")
	fmt.Println("- Default transport capacity: 10 messages")
	fmt.Println("- Default query capacity: 100 messages")
	fmt.Println("- Useful for tuning performance in different scenarios")
	fmt.Println("- Higher capacity: better for bursty traffic, more memory usage")
	fmt.Println("- Lower capacity: less memory usage, potential blocking under load")
}
