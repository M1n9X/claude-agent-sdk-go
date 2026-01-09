package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	claude "github.com/M1n9X/claude-agent-sdk-go"
	"github.com/M1n9X/claude-agent-sdk-go/types"
)

// ConcurrentUsage demonstrates different patterns for concurrent usage of the SDK.
func main() {
	ctx := context.Background()

	fmt.Println("Concurrent Usage Patterns")
	fmt.Println("=========================")
	fmt.Println()

	// Pattern 1: Recommended - One Client per Goroutine
	fmt.Println("Pattern 1: One Client per Goroutine (Recommended)")
	fmt.Println("--------------------------------------------------")
	pattern1OneClientPerGoroutine(ctx)

	fmt.Println()

	// Pattern 2: Shared Client with ConcurrentClient
	fmt.Println("Pattern 2: Shared Client with ConcurrentClient")
	fmt.Println("-----------------------------------------------")
	pattern2SharedConcurrentClient(ctx)

	fmt.Println()

	// Pattern 3: Manual Synchronization (Not Recommended)
	fmt.Println("Pattern 3: Manual Synchronization (Not Recommended)")
	fmt.Println("---------------------------------------------------")
	pattern3ManualSync(ctx)

	fmt.Println()
	fmt.Println("All patterns completed!")
}

// Pattern 1: One Client per Goroutine (Recommended)
// This is the recommended pattern. Each goroutine creates its own client,
// which is the most straightforward and efficient approach.
func pattern1OneClientPerGoroutine(ctx context.Context) {
	opts := types.NewClaudeAgentOptions().
		WithModel("claude-sonnet-4-5-20250929")

	var wg sync.WaitGroup
	numTasks := 3

	for i := 0; i < numTasks; i++ {
		wg.Add(1)
		go func(taskID int) {
			defer wg.Done()

			// Each goroutine creates its own client
			client, err := claude.NewClient(ctx, opts)
			if err != nil {
				log.Printf("Task %d: Failed to create client: %v", taskID, err)
				return
			}
			defer client.Close(ctx)

			if err := client.Connect(ctx); err != nil {
				log.Printf("Task %d: Failed to connect: %v", taskID, err)
				return
			}

			prompt := fmt.Sprintf("What is %d + %d? Answer in one sentence.", taskID, taskID)
			if err := client.Query(ctx, prompt); err != nil {
				log.Printf("Task %d: Query failed: %v", taskID, err)
				return
			}

			fmt.Printf("Task %d: Query sent\n", taskID)

			for msg := range client.ReceiveResponse(ctx) {
				if assistantMsg, ok := msg.(*types.AssistantMessage); ok {
					for _, block := range assistantMsg.Content {
						if textBlock, ok := block.(*types.TextBlock); ok {
							fmt.Printf("Task %d: %s\n", taskID, textBlock.Text)
						}
					}
				}
			}
		}(i)
	}

	wg.Wait()
	fmt.Println("Pattern 1 completed")
}

// Pattern 2: Shared Client with ConcurrentClient
// Use this pattern if you need to share a single client across goroutines.
// The ConcurrentClient wrapper provides thread-safety.
func pattern2SharedConcurrentClient(ctx context.Context) {
	opts := types.NewClaudeAgentOptions().
		WithModel("claude-sonnet-4-5-20250929")

	// Create a concurrent client (thread-safe)
	client, err := claude.NewConcurrentClient(ctx, opts)
	if err != nil {
		log.Printf("Failed to create concurrent client: %v", err)
		return
	}
	defer client.Close(ctx)

	if err := client.Connect(ctx); err != nil {
		log.Printf("Failed to connect: %v", err)
		return
	}

	numTasks := 3

	// Multiple producers enqueue work; a single worker executes queries serially
	// to avoid interleaved responses on the shared session.
	tasks := make(chan int, numTasks)

	// Producers
	var producers sync.WaitGroup
	for i := 0; i < numTasks; i++ {
		producers.Add(1)
		go func(taskID int) {
			defer producers.Done()
			tasks <- taskID
		}(i)
	}

	// Worker
	var worker sync.WaitGroup
	worker.Add(1)
	go func() {
		defer worker.Done()
		for taskID := range tasks {
			prompt := fmt.Sprintf("What is %d * 2? Answer in one sentence.", taskID)

			messages, err := client.QueryAndReceive(ctx, prompt)
			if err != nil {
				log.Printf("Task %d: Query failed: %v", taskID, err)
				continue
			}

			fmt.Printf("Task %d: Query sent (shared client)\n", taskID)
			for msg := range messages {
				if assistantMsg, ok := msg.(*types.AssistantMessage); ok {
					for _, block := range assistantMsg.Content {
						if textBlock, ok := block.(*types.TextBlock); ok {
							fmt.Printf("Task %d: %s\n", taskID, textBlock.Text)
						}
					}
				}
			}
		}
	}()

	// Close task queue after producers finish
	producers.Wait()
	close(tasks)

	// Wait for worker to drain
	worker.Wait()
	fmt.Println("Pattern 2 completed")
}

// Pattern 3: Manual Synchronization (Not Recommended)
// This pattern shows how to manually synchronize access to a regular Client.
// This is more complex and error-prone than using ConcurrentClient.
func pattern3ManualSync(ctx context.Context) {
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

	// Manual mutex for synchronization
	var mu sync.Mutex
	var wg sync.WaitGroup
	numTasks := 3

	for i := 0; i < numTasks; i++ {
		wg.Add(1)
		go func(taskID int) {
			defer wg.Done()

			// Manually lock before using the client
			mu.Lock()
			defer mu.Unlock()

			prompt := fmt.Sprintf("What is %d - 1? Answer in one sentence.", taskID)
			if err := client.Query(ctx, prompt); err != nil {
				log.Printf("Task %d: Query failed: %v", taskID, err)
				return
			}

			fmt.Printf("Task %d: Query sent (manual sync)\n", taskID)

			for msg := range client.ReceiveResponse(ctx) {
				if assistantMsg, ok := msg.(*types.AssistantMessage); ok {
					for _, block := range assistantMsg.Content {
						if textBlock, ok := block.(*types.TextBlock); ok {
							fmt.Printf("Task %d: %s\n", taskID, textBlock.Text)
						}
					}
				}
			}
		}(i)
	}

	wg.Wait()
	fmt.Println("Pattern 3 completed")
}

// Bonus: Query function is naturally concurrent-safe
// The Query function creates a new connection for each call,
// so it's safe to call from multiple goroutines without any synchronization.
func bonusQueryFunctionConcurrency(ctx context.Context) {
	opts := types.NewClaudeAgentOptions().
		WithModel("claude-sonnet-4-5-20250929")

	var wg sync.WaitGroup
	numTasks := 5

	for i := 0; i < numTasks; i++ {
		wg.Add(1)
		go func(taskID int) {
			defer wg.Done()

			prompt := fmt.Sprintf("What is %d squared? Answer in one sentence.", taskID)
			messages, err := claude.Query(ctx, prompt, opts)
			if err != nil {
				log.Printf("Task %d: Query failed: %v", taskID, err)
				return
			}

			for msg := range messages {
				if assistantMsg, ok := msg.(*types.AssistantMessage); ok {
					for _, block := range assistantMsg.Content {
						if textBlock, ok := block.(*types.TextBlock); ok {
							fmt.Printf("Task %d: %s\n", taskID, textBlock.Text)
						}
					}
				}
			}
		}(i)
	}

	wg.Wait()
}

// Performance comparison
func performanceComparison(ctx context.Context) {
	opts := types.NewClaudeAgentOptions().
		WithModel("claude-sonnet-4-5-20250929")

	numTasks := 10

	// Pattern 1: One client per goroutine
	start := time.Now()
	var wg sync.WaitGroup
	for i := 0; i < numTasks; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			client, _ := claude.NewClient(ctx, opts)
			defer client.Close(ctx)
			client.Connect(ctx)
			client.Query(ctx, fmt.Sprintf("Task %d", id))
			for range client.ReceiveResponse(ctx) {
			}
		}(i)
	}
	wg.Wait()
	pattern1Time := time.Since(start)

	// Pattern 2: Shared concurrent client
	start = time.Now()
	client, _ := claude.NewConcurrentClient(ctx, opts)
	defer client.Close(ctx)
	client.Connect(ctx)

	for i := 0; i < numTasks; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			client.Query(ctx, fmt.Sprintf("Task %d", id))
			for range client.ReceiveResponse(ctx) {
			}
		}(i)
	}
	wg.Wait()
	pattern2Time := time.Since(start)

	fmt.Printf("Pattern 1 (one client per goroutine): %v\n", pattern1Time)
	fmt.Printf("Pattern 2 (shared concurrent client): %v\n", pattern2Time)
}
