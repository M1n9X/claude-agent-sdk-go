// Package transport implements HTTP and SSE transport for MCP servers.
package transport

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/M1n9X/claude-agent-sdk-go/internal/log"
	"github.com/M1n9X/claude-agent-sdk-go/internal/mcp"
	"github.com/M1n9X/claude-agent-sdk-go/types"
)

// HTTPTransport implements HTTP-based MCP transport
type HTTPTransport struct {
	url         string
	headers     map[string]string
	client      *http.Client
	messageChan chan []byte
	errChan     chan error
	logger      *log.Logger

	ctx    context.Context
	cancel context.CancelFunc

	// sseMode indicates if this is an SSE (Server-Sent Events) connection
	sseMode bool

	// For SSE: store the HTTP response body to close later
	respBody io.ReadCloser

	// Mutex for operations
	mu sync.RWMutex
}

// NewHTTPTransport creates a new HTTP transport
func NewHTTPTransport(url string, headers map[string]string, logger *log.Logger) *HTTPTransport {
	// Add default headers if not provided
	transHeaders := make(map[string]string)
	if headers != nil {
		for k, v := range headers {
			transHeaders[k] = v
		}
	}

	// Default Content-Type if not set
	if _, ok := transHeaders["Content-Type"]; !ok {
		transHeaders["Content-Type"] = "application/json"
	}

	return &HTTPTransport{
		url:     url,
		headers: transHeaders,
		client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     30 * time.Second,
			},
		},
		messageChan: make(chan []byte, 100),
		errChan:     make(chan error, 10),
		logger:      logger,
		sseMode:     strings.Contains(url, "/sse"),
	}
}

// Connect establishes connection to the MCP server
func (t *HTTPTransport) Connect(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.ctx != nil {
		return fmt.Errorf("already connected")
	}

	t.ctx, t.cancel = context.WithCancel(ctx)
	t.logger.Debug("Connecting to MCP server: %s", t.url)

	// Send initialize request
	initRequest := mcp.NewRequest("initialize", map[string]interface{}{
		"protocolVersion": "0.1.0",
		"capabilities":    map[string]interface{}{},
	})

	initData, err := initRequest.Marshal()
	if err != nil {
		return fmt.Errorf("marshal initialize request: %w", err)
	}

	// Start SSE receiver if in SSE mode
	if t.sseMode {
		go t.sseReceiver()
	}

	// Send initialize request and wait for response
	resp, err := t.sendHTTPRequest("POST", t.url, initData)
	if err != nil {
		return fmt.Errorf("send initialize request: %w", err)
	}

	// Parse initialize response
	var initResponse mcp.Response
	if err := json.Unmarshal(resp, &initResponse); err != nil {
		return fmt.Errorf("parse initialize response: %w", err)
	}

	if initResponse.Error != nil {
		return fmt.Errorf("initialize failed: %s", initResponse.Error.Message)
	}

	t.logger.Debug("Successfully connected to MCP server")
	return nil
}

// sendHTTPRequest sends an HTTP request and returns the response body
func (t *HTTPTransport) sendHTTPRequest(method, url string, body []byte) ([]byte, error) {
	req, err := http.NewRequestWithContext(t.ctx, method, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	// Set headers
	for k, v := range t.headers {
		req.Header.Set(k, v)
	}

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("bad status: %d - %s", resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}

// sseReceiver handles Server-Sent Events connection
func (t *HTTPTransport) sseReceiver() {
	defer close(t.messageChan)

	t.logger.Debug("Starting SSE receiver for: %s", t.url)

	req, err := http.NewRequestWithContext(t.ctx, "GET", t.url, nil)
	if err != nil {
		t.errChan <- fmt.Errorf("create SSE request: %w", err)
		return
	}

	// Set headers
	for k, v := range t.headers {
		req.Header.Set(k, v)
	}
	req.Header.Set("Accept", "text/event-stream")

	resp, err := t.client.Do(req)
	if err != nil {
		t.errChan <- fmt.Errorf("SSE request failed: %w", err)
		return
	}

	// Store response body for cleanup
	t.mu.Lock()
	t.respBody = resp.Body
	t.mu.Unlock()

	defer resp.Body.Close()

	// Check for correct content type
	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/event-stream") {
		t.errChan <- fmt.Errorf("unexpected content type: %s", contentType)
		return
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		select {
		case <-t.ctx.Done():
			return
		default:
		}

		line := scanner.Text()
		t.logger.Debug("SSE data: %s", line)

		// Parse SSE data line (format: "data: {json}")
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			if data != "" {
				select {
				case t.messageChan <- []byte(data):
				default:
					// Drop message if channel is full (backpressure)
					t.logger.Warning("Message channel full, dropping SSE message")
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		t.errChan <- fmt.Errorf("SSE scanner error: %w", err)
	}
}

// Write sends a JSON-RPC request to the MCP server
func (t *HTTPTransport) Write(ctx context.Context, data string) error {
	request, err := mcp.UnmarshalRequest([]byte(data))
	if err != nil {
		// If it's not a valid request, send it as-is
		return t.sendRequest([]byte(data))
	}

	// Build proper JSON-RPC request
	jsonData, err := request.Marshal()
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	return t.sendRequest(jsonData)
}

// sendRequest sends an HTTP request
func (t *HTTPTransport) sendRequest(data []byte) error {
	t.logger.Debug("Sending HTTP request: %s", string(data))

	resp, err := t.sendHTTPRequest("POST", t.url, data)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}

	// Send response to message channel
	select {
	case <-t.ctx.Done():
		return t.ctx.Err()
	case t.messageChan <- resp:
		return nil
	default:
		return fmt.Errorf("message channel full")
	}
}

// ReadMessages returns a channel of incoming JSON-RPC responses
func (t *HTTPTransport) ReadMessages(ctx context.Context) <-chan types.Message {
	// For HTTP transport, we need to parse the JSON messages
	// and convert them to types.Message
	msgChan := make(chan types.Message)

	go func() {
		defer close(msgChan)

		for {
			select {
			case <-ctx.Done():
				return
			case <-t.ctx.Done():
				return
			case data, ok := <-t.messageChan:
				if !ok {
					return
				}

				// Parse JSON-RPC response
				var response mcp.Response
				if err := json.Unmarshal(data, &response); err != nil {
					t.logger.Error("Failed to parse response: %v", err)
					continue
				}

				// Convert to transport message format
				msg := &types.JSONMessage{
					Data: data,
				}

				select {
				case msgChan <- msg:
				case <-ctx.Done():
					return
				case <-t.ctx.Done():
					return
				}
			}
		}
	}()

	return msgChan
}

// OnError stores an error
func (t *HTTPTransport) OnError(err error) {
	select {
	case t.errChan <- err:
	default:
		// Drop error if channel is full
	}
}

// IsReady returns true if transport is ready
func (t *HTTPTransport) IsReady() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.ctx != nil && t.ctx.Err() == nil
}

// GetError returns any error
func (t *HTTPTransport) GetError() error {
	select {
	case err := <-t.errChan:
		return err
	default:
		return nil
	}
}

// Close closes the transport
func (t *HTTPTransport) Close(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.cancel != nil {
		t.cancel()
	}

	if t.sseMode && t.respBody != nil {
		t.respBody.Close()
	}

	close(t.messageChan)
	close(t.errChan)

	return nil
}

// NewHTTPTransportFromConfig creates an HTTP transport from a config
func NewHTTPTransportFromConfig(config types.McpHTTPServerConfig, logger *log.Logger) *HTTPTransport {
	headers := make(map[string]string)
	for k, v := range config.Headers {
		headers[k] = v
	}
	return NewHTTPTransport(config.URL, headers, logger)
}

// NewSSETransportFromConfig creates an SSE transport from a config
func NewSSETransportFromConfig(config types.McpSSEServerConfig, logger *log.Logger) *HTTPTransport {
	headers := make(map[string]string)
	for k, v := range config.Headers {
		headers[k] = v
	}
	url := config.URL
	if !strings.Contains(url, "/") {
		url = url + "/sse"
	}
	return NewHTTPTransport(url, headers, logger)
}
