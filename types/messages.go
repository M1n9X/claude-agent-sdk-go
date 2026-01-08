package types

import (
	"encoding/json"
	"fmt"
)

// SystemMessageSubtype constants for common system message subtypes
const (
	SystemSubtypeInit        = "init"
	SystemSubtypeWarning     = "warning"
	SystemSubtypeError       = "error"
	SystemSubtypeInfo        = "info"
	SystemSubtypeDebug       = "debug"
	SystemSubtypeSessionEnd  = "session_end"
	SystemSubtypeSessionInfo = "session_info"
)

// ContentBlock is an interface for all content block types.
// Content blocks can be text, thinking, tool use, or tool result blocks.
type ContentBlock interface {
	GetType() string
	isContentBlock()
}

// AssistantMessageError represents known assistant error codes.
type AssistantMessageError string

const (
	AssistantMessageErrorAuthenticationFailed AssistantMessageError = "authentication_failed"
	AssistantMessageErrorBilling              AssistantMessageError = "billing_error"
	AssistantMessageErrorRateLimit            AssistantMessageError = "rate_limit"
	AssistantMessageErrorInvalidRequest       AssistantMessageError = "invalid_request"
	AssistantMessageErrorServer               AssistantMessageError = "server_error"
	AssistantMessageErrorUnknown              AssistantMessageError = "unknown"
)

// TextBlock represents a text content block from Claude.
type TextBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// GetType returns the type of the content block.
func (t TextBlock) GetType() string {
	return t.Type
}

func (t TextBlock) isContentBlock() {}

// ThinkingBlock represents a thinking content block from Claude.
// This contains Claude's internal reasoning and signature.
type ThinkingBlock struct {
	Type      string `json:"type"`
	Thinking  string `json:"thinking"`
	Signature string `json:"signature"`
}

// GetType returns the type of the content block.
func (t ThinkingBlock) GetType() string {
	return t.Type
}

func (t ThinkingBlock) isContentBlock() {}

// ToolUseBlock represents a tool use request from Claude.
type ToolUseBlock struct {
	Type  string                 `json:"type"`
	ID    string                 `json:"id"`
	Name  string                 `json:"name"`
	Input map[string]interface{} `json:"input"`
}

// GetType returns the type of the content block.
func (t ToolUseBlock) GetType() string {
	return t.Type
}

func (t ToolUseBlock) isContentBlock() {}

// ToolResultBlock represents the result of a tool execution.
type ToolResultBlock struct {
	Type      string      `json:"type"`
	ToolUseID string      `json:"tool_use_id"`
	Content   interface{} `json:"content,omitempty"`  // Can be string or []map[string]interface{}
	IsError   *bool       `json:"is_error,omitempty"` // Pointer to distinguish between false and not set
}

// GetType returns the type of the content block.
func (t ToolResultBlock) GetType() string {
	return t.Type
}

func (t ToolResultBlock) isContentBlock() {}

// UnmarshalContentBlock unmarshals a JSON content block into the appropriate type.
func UnmarshalContentBlock(data []byte) (ContentBlock, error) {
	var typeCheck struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(data, &typeCheck); err != nil {
		return nil, NewCLIJSONDecodeErrorWithCause("failed to determine content block type", string(data), err)
	}

	switch typeCheck.Type {
	case "text":
		var block TextBlock
		if err := json.Unmarshal(data, &block); err != nil {
			return nil, NewCLIJSONDecodeErrorWithCause("failed to unmarshal text block", string(data), err)
		}
		return &block, nil
	case "thinking":
		var block ThinkingBlock
		if err := json.Unmarshal(data, &block); err != nil {
			return nil, NewCLIJSONDecodeErrorWithCause("failed to unmarshal thinking block", string(data), err)
		}
		return &block, nil
	case "tool_use":
		var block ToolUseBlock
		if err := json.Unmarshal(data, &block); err != nil {
			return nil, NewCLIJSONDecodeErrorWithCause("failed to unmarshal tool_use block", string(data), err)
		}
		return &block, nil
	case "tool_result":
		var block ToolResultBlock
		if err := json.Unmarshal(data, &block); err != nil {
			return nil, NewCLIJSONDecodeErrorWithCause("failed to unmarshal tool_result block", string(data), err)
		}
		return &block, nil
	default:
		return nil, NewMessageParseErrorWithType("unknown content block type", typeCheck.Type)
	}
}

// Message is an interface for all message types from Claude.
type Message interface {
	GetMessageType() string
	ShouldDisplayToUser() bool
	isMessage()
}

// Type-safe accessor methods for different message types.
func (m *UserMessage) AsUser() (*UserMessage, bool) {
	return m, true
}

func (m *AssistantMessage) AsAssistant() (*AssistantMessage, bool) {
	return m, true
}

func (m *SystemMessage) AsSystem() (*SystemMessage, bool) {
	return m, true
}

func (m *ResultMessage) AsResult() (*ResultMessage, bool) {
	return m, true
}

func (m *StreamEvent) AsStreamEvent() (*StreamEvent, bool) {
	return m, true
}

func (m *JSONMessage) AsJSON() (*JSONMessage, bool) {
	return m, true
}

// AsType provides type-safe conversion to specific message types.
func (m *UserMessage) AsType() (user *UserMessage, assistant *AssistantMessage, system *SystemMessage, result *ResultMessage, streamEvent *StreamEvent, jsonMsg *JSONMessage) {
	return m, nil, nil, nil, nil, nil
}

func (m *AssistantMessage) AsType() (user *UserMessage, assistant *AssistantMessage, system *SystemMessage, result *ResultMessage, streamEvent *StreamEvent, jsonMsg *JSONMessage) {
	return nil, m, nil, nil, nil, nil
}

func (m *SystemMessage) AsType() (user *UserMessage, assistant *AssistantMessage, system *SystemMessage, result *ResultMessage, streamEvent *StreamEvent, jsonMsg *JSONMessage) {
	return nil, nil, m, nil, nil, nil
}

func (m *ResultMessage) AsType() (user *UserMessage, assistant *AssistantMessage, system *SystemMessage, result *ResultMessage, streamEvent *StreamEvent, jsonMsg *JSONMessage) {
	return nil, nil, nil, m, nil, nil
}

func (m *StreamEvent) AsType() (user *UserMessage, assistant *AssistantMessage, system *SystemMessage, result *ResultMessage, streamEvent *StreamEvent, jsonMsg *JSONMessage) {
	return nil, nil, nil, nil, m, nil
}

func (m *JSONMessage) AsType() (user *UserMessage, assistant *AssistantMessage, system *SystemMessage, result *ResultMessage, streamEvent *StreamEvent, jsonMsg *JSONMessage) {
	return nil, nil, nil, nil, nil, m
}

// UserMessage represents a message from the user.
type UserMessage struct {
	Type            string      `json:"type"`
	Content         interface{} `json:"content"` // Can be string or []ContentBlock
	ParentToolUseID *string     `json:"parent_tool_use_id,omitempty"`
	UUID            *string     `json:"uuid,omitempty"`
}

// GetMessageType returns the type of the message.
func (m *UserMessage) GetMessageType() string {
	return m.Type
}

// ShouldDisplayToUser returns true for user messages (always display).
func (m *UserMessage) ShouldDisplayToUser() bool {
	return true
}

func (m *UserMessage) isMessage() {}

// JSONMessage represents a raw JSON message for transport.
// This is used for low-level protocol communication where the message
// content is already in JSON format and doesn't need to be re-marshaled.
type JSONMessage struct {
	Data []byte
}

// GetMessageType returns the type of the message.
func (m *JSONMessage) GetMessageType() string {
	return "json"
}

// ShouldDisplayToUser indicates if the message should be displayed to the user.
func (m *JSONMessage) ShouldDisplayToUser() bool {
	return true // JSON messages are used for internal protocol communication
}

func (m *JSONMessage) isMessage() {}

// MarshalJSON returns the JSON data without re-encoding.
func (m *JSONMessage) MarshalJSON() ([]byte, error) {
	return m.Data, nil
}

// UnmarshalJSON implements custom unmarshaling for UserMessage to handle content union type.
func (m *UserMessage) UnmarshalJSON(data []byte) error {
	type Alias UserMessage
	aux := &struct {
		Content json.RawMessage            `json:"content"`
		Message map[string]json.RawMessage `json:"message"` // Handle nested message format from CLI
		UUID    *string                    `json:"uuid"`
		*Alias
	}{
		Alias: (*Alias)(m),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Capture UUID from top-level
	if aux.UUID != nil {
		m.UUID = aux.UUID
	}

	var contentRaw json.RawMessage

	// Check if content is in nested message.content (Claude CLI format)
	if aux.Message != nil {
		if content, ok := aux.Message["content"]; ok {
			contentRaw = content
		}
		// Also extract parent_tool_use_id from nested message if present
		if parentToolUseID, ok := aux.Message["parent_tool_use_id"]; ok {
			var id string
			if err := json.Unmarshal(parentToolUseID, &id); err == nil {
				m.ParentToolUseID = &id
			}
		}
		// Extract uuid from nested message if present
		if uuidRaw, ok := aux.Message["uuid"]; ok {
			var id string
			if err := json.Unmarshal(uuidRaw, &id); err == nil {
				m.UUID = &id
			}
		}
	}

	// Fall back to top-level content if nested not found
	if contentRaw == nil && aux.Content != nil {
		contentRaw = aux.Content
	}

	// If we still don't have content, that's an error
	if contentRaw == nil {
		return fmt.Errorf("missing content field")
	}

	// Try to unmarshal as string first
	var contentStr string
	if err := json.Unmarshal(contentRaw, &contentStr); err == nil {
		m.Content = contentStr
		return nil
	}

	// Try to unmarshal as array of content blocks
	var contentArr []json.RawMessage
	if err := json.Unmarshal(contentRaw, &contentArr); err == nil {
		blocks := make([]ContentBlock, len(contentArr))
		for i, rawBlock := range contentArr {
			block, err := UnmarshalContentBlock(rawBlock)
			if err != nil {
				return err
			}
			blocks[i] = block
		}
		m.Content = blocks
		return nil
	}

	return fmt.Errorf("content must be string or array of content blocks")
}

// AssistantMessage represents a message from Claude assistant.
type AssistantMessage struct {
	Type            string                 `json:"type"`
	Content         []ContentBlock         `json:"content"`
	Model           string                 `json:"model"`
	ParentToolUseID *string                `json:"parent_tool_use_id,omitempty"`
	Error           *AssistantMessageError `json:"error,omitempty"`
}

// GetMessageType returns the type of the message.
func (m *AssistantMessage) GetMessageType() string {
	return m.Type
}

// ShouldDisplayToUser returns true for assistant messages (always display).
func (m *AssistantMessage) ShouldDisplayToUser() bool {
	return true
}

func (m *AssistantMessage) isMessage() {}

// UnmarshalJSON implements custom unmarshaling for AssistantMessage to handle content blocks.
func (m *AssistantMessage) UnmarshalJSON(data []byte) error {
	type Alias AssistantMessage
	aux := &struct {
		Content []json.RawMessage          `json:"content"`
		Message map[string]json.RawMessage `json:"message"` // Handle nested message format from CLI
		*Alias
	}{
		Alias: (*Alias)(m),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	var contentBlocks []json.RawMessage

	// Check if content is in nested message.content (Claude CLI format)
	if aux.Message != nil {
		if contentRaw, ok := aux.Message["content"]; ok {
			var nested []json.RawMessage
			if err := json.Unmarshal(contentRaw, &nested); err == nil {
				contentBlocks = nested
			}
		}
		// Also extract model from nested message if present
		if modelRaw, ok := aux.Message["model"]; ok {
			var model string
			if err := json.Unmarshal(modelRaw, &model); err == nil {
				m.Model = model
			}
		}
		// Extract error field for rate limit/other errors
		if errRaw, ok := aux.Message["error"]; ok {
			var errCode string
			if err := json.Unmarshal(errRaw, &errCode); err == nil && errCode != "" {
				code := AssistantMessageError(errCode)
				m.Error = &code
			}
		}
	}

	// Fall back to top-level content if nested not found
	if contentBlocks == nil && aux.Content != nil {
		contentBlocks = aux.Content
	}

	// Unmarshal content blocks
	m.Content = make([]ContentBlock, len(contentBlocks))
	for i, rawBlock := range contentBlocks {
		block, err := UnmarshalContentBlock(rawBlock)
		if err != nil {
			return err
		}
		m.Content[i] = block
	}

	return nil
}

// MarshalJSON implements custom marshaling for AssistantMessage to handle content blocks.
func (m *AssistantMessage) MarshalJSON() ([]byte, error) {
	type Alias AssistantMessage
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(m),
	})
}

// SystemMessage represents a system message with metadata.
type SystemMessage struct {
	Type      string                 `json:"type"`
	Subtype   string                 `json:"subtype,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Response  map[string]interface{} `json:"response,omitempty"`   // For control_response messages
	Request   map[string]interface{} `json:"request,omitempty"`    // For control_request messages
	RequestID string                 `json:"request_id,omitempty"` // For control_request/control_response messages (top-level field)
}

// GetMessageType returns the type of the message.
func (m *SystemMessage) GetMessageType() string {
	return m.Type
}

func (m *SystemMessage) isMessage() {}

// IsInit returns true if this is a system init message.
func (m *SystemMessage) IsInit() bool {
	return m.Subtype == SystemSubtypeInit
}

// IsWarning returns true if this is a system warning message.
func (m *SystemMessage) IsWarning() bool {
	return m.Subtype == SystemSubtypeWarning
}

// IsError returns true if this is a system error message.
func (m *SystemMessage) IsError() bool {
	return m.Subtype == SystemSubtypeError
}

// IsInfo returns true if this is a system info message.
func (m *SystemMessage) IsInfo() bool {
	return m.Subtype == SystemSubtypeInfo
}

// IsDebug returns true if this is a system debug message.
func (m *SystemMessage) IsDebug() bool {
	return m.Subtype == SystemSubtypeDebug
}

// ShouldDisplayToUser returns true if this system message should be shown to the user.
// By default, init and debug messages are not shown to users.
func (m *SystemMessage) ShouldDisplayToUser() bool {
	return m.Subtype != SystemSubtypeInit && m.Subtype != SystemSubtypeDebug
}

// ResultMessage represents a result message with cost and usage information.
type ResultMessage struct {
	Type             string                 `json:"type"`
	Subtype          string                 `json:"subtype"`
	DurationMs       int                    `json:"duration_ms"`
	DurationAPIMs    int                    `json:"duration_api_ms"`
	IsError          bool                   `json:"is_error"`
	NumTurns         int                    `json:"num_turns"`
	SessionID        string                 `json:"session_id"`
	TotalCostUSD     *float64               `json:"total_cost_usd,omitempty"`
	Usage            map[string]interface{} `json:"usage,omitempty"`
	Result           *string                `json:"result,omitempty"`
	StructuredOutput interface{}            `json:"structured_output,omitempty"`
}

// GetMessageType returns the type of the message.
func (m *ResultMessage) GetMessageType() string {
	return m.Type
}

// ShouldDisplayToUser returns false for result messages (metadata only).
func (m *ResultMessage) ShouldDisplayToUser() bool {
	return false
}

func (m *ResultMessage) isMessage() {}

// StreamEvent represents a stream event for partial message updates during streaming.
type StreamEvent struct {
	Type            string                 `json:"type"`
	UUID            string                 `json:"uuid"`
	SessionID       string                 `json:"session_id"`
	Event           map[string]interface{} `json:"event"` // The raw Anthropic API stream event
	ParentToolUseID *string                `json:"parent_tool_use_id,omitempty"`
}

// GetMessageType returns the type of the message.
func (m *StreamEvent) GetMessageType() string {
	return m.Type
}

// ShouldDisplayToUser returns false for stream events (internal only).
func (m *StreamEvent) ShouldDisplayToUser() bool {
	return false
}

func (m *StreamEvent) isMessage() {}

// UnmarshalMessage unmarshals a JSON message into the appropriate message type.
func UnmarshalMessage(data []byte) (Message, error) {
	var typeCheck struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(data, &typeCheck); err != nil {
		return nil, NewCLIJSONDecodeErrorWithCause("failed to determine message type", string(data), err)
	}

	switch typeCheck.Type {
	case "user":
		var msg UserMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			return nil, NewCLIJSONDecodeErrorWithCause("failed to unmarshal user message", string(data), err)
		}
		return &msg, nil
	case "assistant":
		var msg AssistantMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			return nil, NewCLIJSONDecodeErrorWithCause("failed to unmarshal assistant message", string(data), err)
		}
		return &msg, nil
	case "system", "control_request", "control_response":
		// system, control_request, and control_response are all SystemMessage types
		var msg SystemMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			return nil, NewCLIJSONDecodeErrorWithCause("failed to unmarshal system message", string(data), err)
		}
		return &msg, nil
	case "result":
		var msg ResultMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			return nil, NewCLIJSONDecodeErrorWithCause("failed to unmarshal result message", string(data), err)
		}
		return &msg, nil
	case "stream_event":
		var msg StreamEvent
		if err := json.Unmarshal(data, &msg); err != nil {
			return nil, NewCLIJSONDecodeErrorWithCause("failed to unmarshal stream event", string(data), err)
		}
		return &msg, nil
	default:
		return nil, NewMessageParseErrorWithType("unknown message type", typeCheck.Type)
	}
}

// AsUser attempts to cast a Message to *UserMessage, returning the message and success status.
func AsUser(msg Message) (*UserMessage, bool) {
	if userMsg, ok := msg.(*UserMessage); ok {
		return userMsg, true
	}
	return nil, false
}

// AsAssistant attempts to cast a Message to *AssistantMessage, returning the message and success status.
func AsAssistant(msg Message) (*AssistantMessage, bool) {
	if assistantMsg, ok := msg.(*AssistantMessage); ok {
		return assistantMsg, true
	}
	return nil, false
}

// AsSystem attempts to cast a Message to *SystemMessage, returning the message and success status.
func AsSystem(msg Message) (*SystemMessage, bool) {
	if systemMsg, ok := msg.(*SystemMessage); ok {
		return systemMsg, true
	}
	return nil, false
}

// AsResult attempts to cast a Message to *ResultMessage, returning the message and success status.
func AsResult(msg Message) (*ResultMessage, bool) {
	if resultMsg, ok := msg.(*ResultMessage); ok {
		return resultMsg, true
	}
	return nil, false
}

// AsStreamEvent attempts to cast a Message to *StreamEvent, returning the message and success status.
func AsStreamEvent(msg Message) (*StreamEvent, bool) {
	if streamEventMsg, ok := msg.(*StreamEvent); ok {
		return streamEventMsg, true
	}
	return nil, false
}

// AsJSON attempts to cast a Message to *JSONMessage, returning the message and success status.
func AsJSON(msg Message) (*JSONMessage, bool) {
	if jsonMsg, ok := msg.(*JSONMessage); ok {
		return jsonMsg, true
	}
	return nil, false
}
