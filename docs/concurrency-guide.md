# Concurrency Guide

This guide explains how to use the Claude Agent SDK for Go in concurrent scenarios.

## Design Philosophy: Intentionally Not Thread-Safe

### This is NOT a Bug - It's a Feature

The `Client` type is **intentionally not thread-safe**. This is a deliberate architectural decision based on:

1. **Session Semantics**
2. **Performance Optimization**
3. **Clear Ownership Model**
4. **Python SDK Alignment**
5. **Go Best Practices**

Let's explore why this design is actually **better** than making everything thread-safe.

> **Shared sessions rule:** If you must share a session, keep exactly **one in-flight query/response at a time**. Use `ConcurrentClient.QueryAndReceive` (or `QueryWithContentAndReceive`) so each request is bound to its own response stream. Do **not** let multiple goroutines range over the same `ReceiveResponse` channel.

---

## Why "Not Thread-Safe" is the Right Choice

### 1. Session Semantics: Conversations are Sequential

**The Problem with Thread-Safe Sessions:**

A conversation with Claude is inherently sequential:
```
You: "What files are in this directory?"
Claude: "Here are the files: ..."
You: "Read the first file"
Claude: "The content is: ..."
```

If multiple goroutines could query the same session concurrently:
```
Goroutine 1: "What files are in this directory?"
Goroutine 2: "Read the first file"  // Which first file? Context is lost!
Claude: [confused response]
```

**The Solution:**

Each conversation thread (goroutine) should have its own Client:
```go
// ✅ CORRECT: Each conversation is independent
go func() {
    client, _ := claude.NewClient(ctx, opts)
    defer client.Close(ctx)
    // This goroutine owns this conversation
}()

go func() {
    client, _ := claude.NewClient(ctx, opts)
    defer client.Close(ctx)
    // This goroutine owns a different conversation
}()
```

**Real-World Analogy:**

Think of Client as a phone call. You don't have multiple people simultaneously talking on the same phone call - that would be chaos. Instead, each person makes their own call.

### 2. Performance: Zero Overhead for the Common Case

**With Thread-Safety (Unnecessary Overhead):**
```go
type Client struct {
    mu sync.Mutex  // Every operation locks
    // ...
}

func (c *Client) Query(ctx context.Context, prompt string) error {
    c.mu.Lock()         // Lock overhead
    defer c.mu.Unlock() // Unlock overhead
    // ... actual work ...
}
```

**Without Thread-Safety (Optimal Performance):**
```go
type Client struct {
    // No mutex needed
    // ...
}

func (c *Client) Query(ctx context.Context, prompt string) error {
    // Direct execution, no synchronization overhead
    // ... actual work ...
}
```

**Performance Impact:**
- Mutex operations: ~20-50ns per lock/unlock
- For a typical query with 10 method calls: 200-500ns overhead
- Multiplied by thousands of queries: significant waste
- **99% of users never need this synchronization**

### 3. Clear Ownership: Prevents Subtle Bugs

**With Shared Client (Potential Bugs):**
```go
// ❌ BAD: Shared mutable state
client, _ := claude.NewClient(ctx, opts)

go func() {
    client.Query(ctx, "Task 1")
    // Wait for response...
}()

go func() {
    client.Query(ctx, "Task 2")
    // Which response belongs to which query?
}()
```

**With Owned Client (Bug-Free):**
```go
// ✅ GOOD: Clear ownership
go func() {
    client, _ := claude.NewClient(ctx, opts)
    defer client.Close(ctx)
    client.Query(ctx, "Task 1")
    // This goroutine owns the response
}()

go func() {
    client, _ := claude.NewClient(ctx, opts)
    defer client.Close(ctx)
    client.Query(ctx, "Task 2")
    // This goroutine owns its response
}()
```

**Benefits:**
- ✅ No race conditions possible
- ✅ Clear data flow
- ✅ Easy to reason about
- ✅ Compiler can optimize better

### 4. Python SDK Alignment: Consistent Behavior

The Python SDK's `ClaudeSDKClient` is also not thread-safe:

**Python SDK:**
```python
# Not thread-safe - same design
client = ClaudeSDKClient(options=options)
# Should not be used from multiple threads
```

**Go SDK:**
```go
// Not thread-safe - matching design
client, _ := claude.NewClient(ctx, opts)
// Should not be used from multiple goroutines
```

**Why This Matters:**
- ✅ Users migrating from Python have familiar patterns
- ✅ Documentation translates directly
- ✅ Same mental model across languages
- ✅ Consistent behavior reduces confusion

### 5. Go Best Practices: "Don't communicate by sharing memory"

**Go Proverb:**
> "Don't communicate by sharing memory; share memory by communicating."

**Anti-Pattern (Shared Client):**
```go
// ❌ Sharing memory (Client) between goroutines
client, _ := claude.NewClient(ctx, opts)
go func() { client.Query(ctx, "Task 1") }()
go func() { client.Query(ctx, "Task 2") }()
```

**Go Idiom (Independent Clients):**
```go
// ✅ Each goroutine owns its resources
go func() {
    client, _ := claude.NewClient(ctx, opts)
    defer client.Close(ctx)
    // Own this Client
}()
```

**Or Communicate via Channels:**
```go
// ✅ Communicate via channels
tasks := make(chan string, 10)
results := make(chan string, 10)

// Worker goroutine
go func() {
    client, _ := claude.NewClient(ctx, opts)
    defer client.Close(ctx)
    
    for task := range tasks {
        // Process task
        results <- result
    }
}()
```

---

## When "Not Thread-Safe" Becomes a Problem

### Rare Case: Truly Shared Session

**Scenario:** Multiple goroutines need to interact with the **same conversation session**.

**Example:**
```go
// Multiple UI components updating the same chat session
// Component 1: User types message
// Component 2: System adds context
// Component 3: Plugin adds information
// All need to interact with the same conversation
```

**Solution:** Use `ConcurrentClient`

```go
client, _ := claude.NewConcurrentClient(ctx, opts)
defer client.Close(ctx)

// Now safe from multiple goroutines
go component1.SendMessage(client)
go component2.AddContext(client)
go component3.AddInfo(client)
```

**Important:** This is a **rare case**. Most applications don't need this.

---

## Overview

The SDK provides three approaches for concurrent usage:

1. **One Client per Goroutine** (Recommended) - Each goroutine creates its own client
2. **ConcurrentClient** - Thread-safe wrapper for shared client access
3. **Query Function** - Naturally concurrent-safe for one-shot queries

## Thread Safety Model

### Client (Not Thread-Safe by Default)

The `Client` type is **not thread-safe** by design, matching the Python SDK behavior:

```go
// ❌ NOT SAFE: Multiple goroutines using the same client
client, _ := claude.NewClient(ctx, opts)
go func() { client.Query(ctx, "Task 1") }()
go func() { client.Query(ctx, "Task 2") }() // Race condition!
```

**Why?**
- Each client represents a stateful session
- Sessions are inherently sequential
- Most use cases don't need concurrent access
- Matches Python SDK design

### Internal Thread Safety

The SDK uses internal synchronization for:
- ✅ Connection state (`connected` flag)
- ✅ Message routing
- ✅ Control protocol
- ✅ Hook callbacks

But **not** for:
- ❌ Query/response cycles
- ❌ Session state

## Pattern 1: One Client per Goroutine (Recommended)

This is the **recommended pattern** for most use cases.

### Why Recommended?
- ✅ Simple and straightforward
- ✅ No synchronization overhead
- ✅ Each goroutine has independent state
- ✅ Best performance
- ✅ Matches Go idioms

### Example (serialized responses)

```go
package main

import (
    "context"
    "fmt"
    "sync"
    
    claude "github.com/M1n9X/claude-agent-sdk-go"
    "github.com/M1n9X/claude-agent-sdk-go/types"
)

func main() {
    ctx := context.Background()
    opts := types.NewClaudeAgentOptions().
        WithModel("claude-sonnet-4-5-20250929")
    
    var wg sync.WaitGroup
    
    // Launch 10 concurrent tasks
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(taskID int) {
            defer wg.Done()
            
            // Each goroutine creates its own client
            client, err := claude.NewClient(ctx, opts)
            if err != nil {
                fmt.Printf("Task %d failed: %v\n", taskID, err)
                return
            }
            defer client.Close(ctx)
            
            // Connect and query
            if err := client.Connect(ctx); err != nil {
                fmt.Printf("Task %d connect failed: %v\n", taskID, err)
                return
            }
            
            if err := client.Query(ctx, fmt.Sprintf("What is %d + %d?", taskID, taskID)); err != nil {
                fmt.Printf("Task %d query failed: %v\n", taskID, err)
                return
            }
            
            // Process responses
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
}
```

### When to Use
- ✅ Independent tasks
- ✅ Parallel processing
- ✅ Batch operations
- ✅ Worker pools

## Pattern 2: ConcurrentClient (Thread-Safe Wrapper)

Use this when you need to share a single client across goroutines.

### Why Use This?
- ✅ Share session state
- ✅ Reuse connection
- ✅ Centralized configuration
- ✅ Thread-safe by design

### Example

```go
package main

import (
    "context"
    "fmt"
    "sync"
    
    claude "github.com/M1n9X/claude-agent-sdk-go"
    "github.com/M1n9X/claude-agent-sdk-go/types"
)

func main() {
    ctx := context.Background()
    opts := types.NewClaudeAgentOptions().
        WithModel("claude-sonnet-4-5-20250929")
    
    // Create a thread-safe client
    client, err := claude.NewConcurrentClient(ctx, opts)
    if err != nil {
        panic(err)
    }
    defer client.Close(ctx)
    
    if err := client.Connect(ctx); err != nil {
        panic(err)
    }
    
    // Producers enqueue tasks
    tasks := make(chan int, 10)
    go func() {
        defer close(tasks)
        for i := 0; i < 10; i++ {
            tasks <- i
        }
    }()

    // Single worker executes queries one-at-a-time to avoid response interleaving
    for taskID := range tasks {
        messages, err := client.QueryAndReceive(ctx, fmt.Sprintf("Task %d", taskID))
        if err != nil {
            fmt.Printf("Task %d failed: %v\n", taskID, err)
            continue
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
    }
}
```

### When to Use
- ✅ Shared session required
- ✅ Connection reuse important
- ✅ Centralized state management
- ⚠️ Note: Operations are serialized (one at a time)

### Performance Considerations

`ConcurrentClient` serializes all operations. Use `QueryAndReceive` / `QueryWithContentAndReceive`
to bind each request to its response and prevent interleaving.

```go
// These will execute one at a time, not in parallel
go client.Query(ctx, "Task 1") // Executes first
go client.Query(ctx, "Task 2") // Waits for Task 1
go client.Query(ctx, "Task 3") // Waits for Task 2
```

For true parallelism, use Pattern 1 (one client per goroutine).

## Pattern 3: Query Function (Naturally Concurrent-Safe)

The `Query()` function is naturally concurrent-safe.

### Why Safe?
- Each call creates a new connection
- No shared state between calls
- Independent lifecycle

### Example

```go
package main

import (
    "context"
    "fmt"
    "sync"
    
    claude "github.com/M1n9X/claude-agent-sdk-go"
    "github.com/M1n9X/claude-agent-sdk-go/types"
)

func main() {
    ctx := context.Background()
    opts := types.NewClaudeAgentOptions().
        WithModel("claude-sonnet-4-5-20250929")
    
    var wg sync.WaitGroup
    
    // Query function is naturally concurrent-safe
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(taskID int) {
            defer wg.Done()
            
            // Each call creates its own connection
            messages, err := claude.Query(ctx, fmt.Sprintf("What is %d squared?", taskID), opts)
            if err != nil {
                fmt.Printf("Task %d failed: %v\n", taskID, err)
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
```

### When to Use
- ✅ One-shot queries
- ✅ Stateless operations
- ✅ Simple concurrent tasks
- ✅ No session management needed

## Comparison

| Pattern | Thread-Safe | Performance | Use Case |
|---------|-------------|-------------|----------|
| One Client per Goroutine | ✅ (isolated) | ⭐⭐⭐⭐⭐ Best | Independent tasks |
| ConcurrentClient | ✅ (synchronized) | ⭐⭐⭐ Good | Shared session |
| Query Function | ✅ (isolated) | ⭐⭐⭐⭐ Very Good | One-shot queries |

## Best Practices

### 1. Choose the Right Pattern

```go
// ✅ Good: Independent tasks
for i := 0; i < 10; i++ {
    go func(id int) {
        client, _ := claude.NewClient(ctx, opts)
        defer client.Close(ctx)
        // Use client
    }(i)
}

// ✅ Good: Shared session
client, _ := claude.NewConcurrentClient(ctx, opts)
defer client.Close(ctx)
for i := 0; i < 10; i++ {
    go func(id int) {
        client.Query(ctx, fmt.Sprintf("Task %d", id))
    }(i)
}

// ✅ Good: One-shot queries
for i := 0; i < 10; i++ {
    go func(id int) {
        claude.Query(ctx, fmt.Sprintf("Task %d", id), opts)
    }(i)
}
```

### 2. Always Clean Up

```go
// ✅ Good: Defer cleanup
client, _ := claude.NewClient(ctx, opts)
defer client.Close(ctx)

// ❌ Bad: Forgot to close
client, _ := claude.NewClient(ctx, opts)
// Memory leak!
```

### 3. Use Context for Cancellation

```go
// ✅ Good: Timeout context
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

messages, _ := claude.Query(ctx, "Long task", opts)
```

### 4. Handle Errors Properly

```go
// ✅ Good: Check all errors
client, err := claude.NewClient(ctx, opts)
if err != nil {
    if types.IsCLINotFoundError(err) {
        log.Fatal("Please install Claude CLI")
    }
    log.Fatal(err)
}
defer client.Close(ctx)

if err := client.Connect(ctx); err != nil {
    log.Fatal(err)
}
```

### 5. Don't Share Regular Client

```go
// ❌ BAD: Race condition
client, _ := claude.NewClient(ctx, opts)
go func() { client.Query(ctx, "Task 1") }()
go func() { client.Query(ctx, "Task 2") }() // RACE!

// ✅ GOOD: Use ConcurrentClient
client, _ := claude.NewConcurrentClient(ctx, opts)
go func() { client.Query(ctx, "Task 1") }()
go func() { client.Query(ctx, "Task 2") }() // Safe!

// ✅ BETTER: One client per goroutine
go func() {
    client, _ := claude.NewClient(ctx, opts)
    defer client.Close(ctx)
    client.Query(ctx, "Task 1")
}()
go func() {
    client, _ := claude.NewClient(ctx, opts)
    defer client.Close(ctx)
    client.Query(ctx, "Task 2")
}()
```

## Performance Tips

### 1. Connection Pooling

For high-throughput scenarios, consider a worker pool:

```go
type Worker struct {
    client *claude.Client
}

func NewWorkerPool(ctx context.Context, size int, opts *types.ClaudeAgentOptions) []*Worker {
    workers := make([]*Worker, size)
    for i := 0; i < size; i++ {
        client, _ := claude.NewClient(ctx, opts)
        client.Connect(ctx)
        workers[i] = &Worker{client: client}
    }
    return workers
}

// Use workers from pool
workers := NewWorkerPool(ctx, 10, opts)
defer func() {
    for _, w := range workers {
        w.client.Close(ctx)
    }
}()

// Distribute tasks to workers
for i, task := range tasks {
    worker := workers[i%len(workers)]
    go worker.client.Query(ctx, task)
}
```

### 2. Batch Processing

```go
// Process in batches
batchSize := 10
for i := 0; i < len(tasks); i += batchSize {
    end := i + batchSize
    if end > len(tasks) {
        end = len(tasks)
    }
    
    batch := tasks[i:end]
    var wg sync.WaitGroup
    
    for _, task := range batch {
        wg.Add(1)
        go func(t string) {
            defer wg.Done()
            claude.Query(ctx, t, opts)
        }(task)
    }
    
    wg.Wait() // Wait for batch to complete
}
```

### 3. Rate Limiting

```go
import "golang.org/x/time/rate"

limiter := rate.NewLimiter(10, 1) // 10 requests per second

for _, task := range tasks {
    limiter.Wait(ctx) // Rate limit
    go claude.Query(ctx, task, opts)
}
```

## Common Pitfalls

### 1. Forgetting to Close

```go
// ❌ BAD: Memory leak
for i := 0; i < 1000; i++ {
    client, _ := claude.NewClient(ctx, opts)
    // Forgot to close!
}

// ✅ GOOD: Always close
for i := 0; i < 1000; i++ {
    client, _ := claude.NewClient(ctx, opts)
    defer client.Close(ctx)
}
```

### 2. Sharing Regular Client

```go
// ❌ BAD: Race condition
client, _ := claude.NewClient(ctx, opts)
for i := 0; i < 10; i++ {
    go client.Query(ctx, fmt.Sprintf("Task %d", i))
}

// ✅ GOOD: Use ConcurrentClient or separate clients
```

### 3. Not Handling Context Cancellation

```go
// ❌ BAD: Ignores cancellation
messages, _ := claude.Query(context.Background(), "Task", opts)

// ✅ GOOD: Respects cancellation
ctx, cancel := context.WithCancel(context.Background())
defer cancel()
messages, _ := claude.Query(ctx, "Task", opts)
```

## Examples

See [examples/advanced/concurrent_usage](../../examples/advanced/concurrent_usage/main.go) for complete working examples of all patterns.

## Summary

| Scenario | Recommended Pattern |
|----------|---------------------|
| Independent parallel tasks | One Client per Goroutine |
| Shared session state | ConcurrentClient |
| One-shot queries | Query Function |
| High throughput | Worker Pool |
| Rate limiting | Query Function + rate.Limiter |

**Default recommendation:** Use **one client per goroutine** unless you have a specific reason to share a client.
