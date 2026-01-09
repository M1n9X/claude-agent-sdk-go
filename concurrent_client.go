package claude

import (
	"context"
	"sync"

	"github.com/M1n9X/claude-agent-sdk-go/types"
)

// ConcurrentClient is a thread-safe wrapper around Client.
// It allows multiple goroutines to safely use the same client instance.
//
// Note: This is typically not needed. The recommended pattern is to create
// a separate Client instance for each goroutine. Use this only if you have
// a specific need to share a single client across goroutines.
//
// Example usage:
//
//	client, _ := claude.NewConcurrentClient(ctx, opts)
//	defer client.Close(ctx)
//
//	client.Connect(ctx)
//
//	// Safe to call from multiple goroutines
//	var wg sync.WaitGroup
//	for i := 0; i < 10; i++ {
//	    wg.Add(1)
//	    go func(id int) {
//	        defer wg.Done()
//	        client.Query(ctx, fmt.Sprintf("Task %d", id))
//	        for msg := range client.ReceiveResponse(ctx) {
//	            // Process messages
//	        }
//	    }(i)
//	}
//	wg.Wait()
type ConcurrentClient struct {
	client *Client
	mu     sync.Mutex // protects client lifecycle calls
	reqMu  sync.Mutex // serializes query/response cycles
}

// NewConcurrentClient creates a new thread-safe client.
//
// This wraps a regular Client with a mutex to provide thread-safety.
// All operations are serialized, so concurrent calls will be executed
// one at a time.
//
// Parameters:
//   - ctx: Parent context for the client lifecycle
//   - options: Configuration options (nil uses defaults)
//
// Returns:
//   - A new ConcurrentClient instance
//   - An error if the CLI cannot be found or options are invalid
func NewConcurrentClient(ctx context.Context, options *types.ClaudeAgentOptions) (*ConcurrentClient, error) {
	client, err := NewClient(ctx, options)
	if err != nil {
		return nil, err
	}

	return &ConcurrentClient{
		client: client,
	}, nil
}

// Connect establishes a connection to Claude Code CLI.
// This method is thread-safe.
func (c *ConcurrentClient) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.client.Connect(ctx)
}

// Query sends a prompt to Claude in the current session.
// This method is thread-safe. Concurrent calls will be serialized.
func (c *ConcurrentClient) Query(ctx context.Context, prompt string) error {
	// Kept for backward compatibility, but does not coordinate ReceiveResponse.
	// Prefer QueryAndReceive for shared-session safety.
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.client.Query(ctx, prompt)
}

// QueryWithContent sends a structured content query to Claude.
// This method is thread-safe. Concurrent calls will be serialized.
func (c *ConcurrentClient) QueryWithContent(ctx context.Context, content interface{}) error {
	// Kept for backward compatibility, but does not coordinate ReceiveResponse.
	// Prefer QueryWithContentAndReceive for shared-session safety.
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.client.QueryWithContent(ctx, content)
}

// ReceiveResponse returns a channel of response messages from Claude.
// This method is thread-safe, but note that only one goroutine should
// consume from the returned channel.
func (c *ConcurrentClient) ReceiveResponse(ctx context.Context) <-chan types.Message {
	// Note: We don't lock here because ReceiveResponse itself is safe to call
	// concurrently (it just returns a channel). The underlying client handles
	// the channel management safely.
	return c.client.ReceiveResponse(ctx)
}

// QueryAndReceive sends a prompt and returns a dedicated channel for its response.
// The entire query/response cycle is serialized so responses cannot interleave
// across goroutines. Next callers will block until this response completes.
func (c *ConcurrentClient) QueryAndReceive(ctx context.Context, prompt string) (<-chan types.Message, error) {
	c.reqMu.Lock()

	if err := c.client.Query(ctx, prompt); err != nil {
		c.reqMu.Unlock()
		return nil, err
	}

	upstream := c.client.ReceiveResponse(ctx)
	out := make(chan types.Message, 10)

	go func() {
		defer close(out)
		defer c.reqMu.Unlock()

		for msg := range upstream {
			out <- msg
			if _, ok := msg.(*types.ResultMessage); ok {
				return
			}
		}
	}()

	return out, nil
}

// QueryWithContentAndReceive is the structured-content variant of QueryAndReceive.
func (c *ConcurrentClient) QueryWithContentAndReceive(ctx context.Context, content interface{}) (<-chan types.Message, error) {
	c.reqMu.Lock()

	if err := c.client.QueryWithContent(ctx, content); err != nil {
		c.reqMu.Unlock()
		return nil, err
	}

	upstream := c.client.ReceiveResponse(ctx)
	out := make(chan types.Message, 10)

	go func() {
		defer close(out)
		defer c.reqMu.Unlock()

		for msg := range upstream {
			out <- msg
			if _, ok := msg.(*types.ResultMessage); ok {
				return
			}
		}
	}()

	return out, nil
}

// Close gracefully terminates the Claude session.
// This method is thread-safe.
func (c *ConcurrentClient) Close(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.client.Close(ctx)
}

// IsConnected returns true if the client is currently connected.
// This method is thread-safe.
func (c *ConcurrentClient) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.client.IsConnected()
}

// Interrupt sends an interrupt request to Claude.
// This method is thread-safe.
func (c *ConcurrentClient) Interrupt(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.client.Interrupt(ctx)
}

// RewindFiles rewinds tracked files to the state at the specified checkpoint.
// This method is thread-safe.
func (c *ConcurrentClient) RewindFiles(ctx context.Context, userMessageID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.client.RewindFiles(ctx, userMessageID)
}

// UnderlyingClient returns the underlying non-thread-safe Client.
// Use this only if you need direct access and can guarantee thread-safety yourself.
func (c *ConcurrentClient) UnderlyingClient() *Client {
	return c.client
}
