// Package mcp implements the Model Context Protocol (MCP) for Claude Code.
// It handles JSON-RPC 2.0 messaging for tool communication.
package mcp

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Request represents a JSON-RPC 2.0 request.
type Request struct {
	JsonRpc string                 `json:"jsonrpc"`
	ID      interface{}            `json:"id,omitempty"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params,omitempty"`
}

// Response represents a JSON-RPC 2.0 response.
type Response struct {
	JsonRpc string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *Error      `json:"error,omitempty"`
}

// Error represents a JSON-RPC 2.0 error.
type Error struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Error codes as defined in JSON-RPC 2.0 specification.
const (
	ErrorCodeParseError     = -32700
	ErrorCodeInvalidRequest = -32600
	ErrorCodeMethodNotFound = -32601
	ErrorCodeInvalidParams  = -32602
	ErrorCodeInternalError  = -32603
)

// NewRequest creates a new JSON-RPC request with a unique ID.
func NewRequest(method string, params map[string]interface{}) *Request {
	return &Request{
		JsonRpc: "2.0",
		ID:      uuid.New().String(),
		Method:  method,
		Params:  params,
	}
}

// NewRequestWithID creates a new JSON-RPC request with a specific ID.
func NewRequestWithID(id interface{}, method string, params map[string]interface{}) *Request {
	return &Request{
		JsonRpc: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}
}

// NewSuccessResponse creates a successful JSON-RPC response.
func NewSuccessResponse(id interface{}, result interface{}) *Response {
	return &Response{
		JsonRpc: "2.0",
		ID:      id,
		Result:  result,
	}
}

// NewErrorResponse creates an error JSON-RPC response.
func NewErrorResponse(id interface{}, code int, message string, data ...interface{}) *Response {
	err := &Error{
		Code:    code,
		Message: message,
	}

	if len(data) > 0 {
		err.Data = data[0]
	}

	return &Response{
		JsonRpc: "2.0",
		ID:      id,
		Error:   err,
	}
}

// NewParseError creates a parse error response.
func NewParseError(id interface{}, message string) *Response {
	return NewErrorResponse(id, ErrorCodeParseError, message)
}

// NewInvalidRequest creates an invalid request error response.
func NewInvalidRequest(id interface{}, message string) *Response {
	return NewErrorResponse(id, ErrorCodeInvalidRequest, message)
}

// NewMethodNotFound creates a method not found error response.
func NewMethodNotFound(id interface{}, method string) *Response {
	return NewErrorResponse(id, ErrorCodeMethodNotFound,
		fmt.Sprintf("Method not found: %s", method))
}

// NewInvalidParams creates an invalid params error response.
func NewInvalidParams(id interface{}, message string) *Response {
	return NewErrorResponse(id, ErrorCodeInvalidParams, message)
}

// NewInternalError creates an internal error response.
func NewInternalError(id interface{}, message string, data ...interface{}) *Response {
	return NewErrorResponse(id, ErrorCodeInternalError, message, data)
}

// Marshal serializes a Request to JSON.
func (r *Request) Marshal() ([]byte, error) {
	data, err := json.Marshal(r)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}
	return data, nil
}

// Marshal serializes a Response to JSON.
func (r *Response) Marshal() ([]byte, error) {
	data, err := json.Marshal(r)
	if err != nil {
		return nil, fmt.Errorf("marshal response: %w", err)
	}
	return data, nil
}

// UnmarshalRequest deserializes a JSON-RPC request.
func UnmarshalRequest(data []byte) (*Request, error) {
	var req Request
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, fmt.Errorf("unmarshal request: %w", err)
	}
	return &req, nil
}

// UnmarshalResponse deserializes a JSON-RPC response.
func UnmarshalResponse(data []byte) (*Response, error) {
	var resp Response
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}
	return &resp, nil
}

// IsNotification returns true if the request is a notification (no ID).
func (r *Request) IsNotification() bool {
	return r.ID == nil
}

// HasError returns true if the response contains an error.
func (r *Response) HasError() bool {
	return r.Error != nil
}

// GetError returns the error if present.
func (r *Response) GetError() *Error {
	return r.Error
}

// RequestIDGenerator generates unique request IDs.
type RequestIDGenerator interface {
	Generate() interface{}
}

// UUIDGenerator generates UUID-based request IDs.
type UUIDGenerator struct {
	mu sync.Mutex
}

// Generate generates a new UUID string.
func (g *UUIDGenerator) Generate() interface{} {
	g.mu.Lock()
	defer g.mu.Unlock()
	return uuid.New().String()
}

// IncrementingIDGenerator generates incrementing integer IDs.
type IncrementingIDGenerator struct {
	mu  sync.Mutex
	id  int64
}

// NewIncrementingIDGenerator creates a new incrementing ID generator.
func NewIncrementingIDGenerator() *IncrementingIDGenerator {
	return &IncrementingIDGenerator{}
}

// Generate generates the next ID.
func (g *IncrementingIDGenerator) Generate() interface{} {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.id++
	return g.id
}

// TimestampedIDGenerator generates timestamp-based IDs.
type TimestampedIDGenerator struct {
	mu sync.Mutex
}

// Generate generates a timestamp-based ID.
func (g *TimestampedIDGenerator) Generate() interface{} {
	g.mu.Lock()
	defer g.mu.Unlock()
	return time.Now().UnixNano()
}
