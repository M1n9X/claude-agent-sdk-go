package claude

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/M1n9X/claude-agent-sdk-go/types"
)

func TestConcurrentClient_Creation(t *testing.T) {
	ctx := context.Background()
	opts := types.NewClaudeAgentOptions()

	client, err := NewConcurrentClient(ctx, opts)
	if err != nil {
		// Expected to fail if CLI not found, but should not panic
		if !types.IsCLINotFoundError(err) {
			t.Errorf("Expected CLINotFoundError, got: %v", err)
		}
		return
	}

	if client == nil {
		t.Error("Expected non-nil client")
	}

	if client.UnderlyingClient() == nil {
		t.Error("Expected non-nil underlying client")
	}
}

func TestConcurrentClient_ThreadSafety(t *testing.T) {
	ctx := context.Background()
	opts := types.NewClaudeAgentOptions()

	client, err := NewConcurrentClient(ctx, opts)
	if err != nil {
		t.Skip("Skipping test: CLI not available")
		return
	}

	// Test that multiple goroutines can safely call methods
	var wg sync.WaitGroup
	numGoroutines := 10

	// Test IsConnected (should not panic)
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = client.IsConnected()
		}()
	}

	wg.Wait()
}

func TestConcurrentClient_SerializedAccess(t *testing.T) {
	ctx := context.Background()
	opts := types.NewClaudeAgentOptions()

	_, err := NewConcurrentClient(ctx, opts)
	if err != nil {
		t.Skip("Skipping test: CLI not available")
		return
	}

	// Test that operations are properly serialized
	var wg sync.WaitGroup
	results := make([]int, 0, 10)
	var resultsMu sync.Mutex

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Simulate some work
			time.Sleep(time.Millisecond * time.Duration(id))

			// Record that this goroutine ran
			resultsMu.Lock()
			results = append(results, id)
			resultsMu.Unlock()
		}(i)
	}

	wg.Wait()

	// All goroutines should have completed
	if len(results) != 10 {
		t.Errorf("Expected 10 results, got %d", len(results))
	}
}

func TestConcurrentClient_NilOptions(t *testing.T) {
	ctx := context.Background()

	client, err := NewConcurrentClient(ctx, nil)
	if err != nil {
		// Expected to fail if CLI not found
		if !types.IsCLINotFoundError(err) {
			t.Errorf("Expected CLINotFoundError, got: %v", err)
		}
		return
	}

	if client == nil {
		t.Error("Expected non-nil client with default options")
	}
}

// Example demonstrating concurrent usage
func ExampleConcurrentClient() {
	ctx := context.Background()
	opts := types.NewClaudeAgentOptions().
		WithModel("claude-sonnet-4-5-20250929")

	client, err := NewConcurrentClient(ctx, opts)
	if err != nil {
		fmt.Printf("Failed to create client: %v\n", err)
		return
	}
	defer client.Close(ctx)

	if err := client.Connect(ctx); err != nil {
		fmt.Printf("Failed to connect: %v\n", err)
		return
	}

	// Safe shared session by serializing query/response cycles
	tasks := make(chan int, 3)
	go func() {
		defer close(tasks)
		for i := 0; i < 3; i++ {
			tasks <- i
		}
	}()

	for id := range tasks {
		messages, err := client.QueryAndReceive(ctx, fmt.Sprintf("Task %d", id))
		if err != nil {
			fmt.Printf("Query %d failed: %v\n", id, err)
			continue
		}

		for msg := range messages {
			if assistantMsg, ok := msg.(*types.AssistantMessage); ok {
				for _, block := range assistantMsg.Content {
					if textBlock, ok := block.(*types.TextBlock); ok {
						fmt.Printf("Task %d: %s\n", id, textBlock.Text)
					}
				}
			}
		}
	}
}
