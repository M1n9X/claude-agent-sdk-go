package types

import (
	"errors"
	"fmt"
)

// CLINotFoundError indicates that the Claude Code CLI binary could not be found.
// This typically occurs when the CLI is not installed or not in PATH.
type CLINotFoundError struct {
	Message string
	Cause   error
}

// Error returns the error message, implementing the error interface.
func (e *CLINotFoundError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

// Is checks if the target error is a CLINotFoundError.
func (e *CLINotFoundError) Is(target error) bool {
	_, ok := target.(*CLINotFoundError)
	return ok
}

// Unwrap returns the wrapped error.
func (e *CLINotFoundError) Unwrap() error {
	return e.Cause
}

// NewCLINotFoundError creates a new CLINotFoundError with the given message.
func NewCLINotFoundError(message string) *CLINotFoundError {
	return &CLINotFoundError{Message: message}
}

// NewCLINotFoundErrorWithCause creates a new CLINotFoundError with the given message and cause.
func NewCLINotFoundErrorWithCause(message string, cause error) *CLINotFoundError {
	return &CLINotFoundError{Message: message, Cause: cause}
}

// CLIConnectionError indicates a failure to connect to the Claude Code CLI process.
// This can occur due to subprocess startup failures, pipe creation errors, or
// communication protocol issues.
type CLIConnectionError struct {
	Message string
	Cause   error
}

// Error returns the error message, implementing the error interface.
func (e *CLIConnectionError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

// Is checks if the target error is a CLIConnectionError.
func (e *CLIConnectionError) Is(target error) bool {
	_, ok := target.(*CLIConnectionError)
	return ok
}

// Unwrap returns the wrapped error.
func (e *CLIConnectionError) Unwrap() error {
	return e.Cause
}

// NewCLIConnectionError creates a new CLIConnectionError with the given message.
func NewCLIConnectionError(message string) *CLIConnectionError {
	return &CLIConnectionError{Message: message}
}

// NewCLIConnectionErrorWithCause creates a new CLIConnectionError with the given message and cause.
func NewCLIConnectionErrorWithCause(message string, cause error) *CLIConnectionError {
	return &CLIConnectionError{Message: message, Cause: cause}
}

// ProcessError indicates an error with the Claude Code CLI subprocess.
// This includes unexpected termination, non-zero exit codes, or signal interruption.
type ProcessError struct {
	Message  string
	ExitCode int
	Cause    error
}

// Error returns the error message, implementing the error interface.
func (e *ProcessError) Error() string {
	msg := e.Message
	if e.ExitCode != 0 {
		msg = fmt.Sprintf("%s (exit code: %d)", msg, e.ExitCode)
	}
	if e.Cause != nil {
		msg = msg + ": " + e.Cause.Error()
	}
	return msg
}

// Is checks if the target error is a ProcessError.
func (e *ProcessError) Is(target error) bool {
	_, ok := target.(*ProcessError)
	return ok
}

// Unwrap returns the wrapped error.
func (e *ProcessError) Unwrap() error {
	return e.Cause
}

// NewProcessError creates a new ProcessError with the given message.
func NewProcessError(message string) *ProcessError {
	return &ProcessError{Message: message}
}

// NewProcessErrorWithCode creates a new ProcessError with the given message and exit code.
func NewProcessErrorWithCode(message string, exitCode int) *ProcessError {
	return &ProcessError{Message: message, ExitCode: exitCode}
}

// NewProcessErrorWithCause creates a new ProcessError with the given message and cause.
func NewProcessErrorWithCause(message string, cause error) *ProcessError {
	return &ProcessError{Message: message, Cause: cause}
}

// CLIJSONDecodeError indicates a failure to parse JSON data from the CLI.
// This can occur when the CLI sends malformed JSON or when the JSON structure
// doesn't match the expected schema.
type CLIJSONDecodeError struct {
	Message string
	Raw     string // The raw JSON that failed to parse
	Cause   error
}

// Error returns the error message, implementing the error interface.
func (e *CLIJSONDecodeError) Error() string {
	msg := e.Message
	if e.Raw != "" {
		// Truncate raw data if too long
		rawSnippet := e.Raw
		if len(rawSnippet) > 100 {
			rawSnippet = rawSnippet[:100] + "..."
		}
		msg = fmt.Sprintf("%s (raw: %s)", msg, rawSnippet)
	}
	if e.Cause != nil {
		msg = msg + ": " + e.Cause.Error()
	}
	return msg
}

// Is checks if the target error is a CLIJSONDecodeError.
func (e *CLIJSONDecodeError) Is(target error) bool {
	_, ok := target.(*CLIJSONDecodeError)
	return ok
}

// Unwrap returns the wrapped error.
func (e *CLIJSONDecodeError) Unwrap() error {
	return e.Cause
}

// NewCLIJSONDecodeError creates a new CLIJSONDecodeError with the given message.
func NewCLIJSONDecodeError(message string) *CLIJSONDecodeError {
	return &CLIJSONDecodeError{Message: message}
}

// NewCLIJSONDecodeErrorWithRaw creates a new CLIJSONDecodeError with the given message and raw JSON.
func NewCLIJSONDecodeErrorWithRaw(message string, raw string) *CLIJSONDecodeError {
	return &CLIJSONDecodeError{Message: message, Raw: raw}
}

// NewCLIJSONDecodeErrorWithCause creates a new CLIJSONDecodeError with the given message, raw JSON, and cause.
func NewCLIJSONDecodeErrorWithCause(message string, raw string, cause error) *CLIJSONDecodeError {
	return &CLIJSONDecodeError{Message: message, Raw: raw, Cause: cause}
}

// JSONDecodeError indicates a failure to parse JSON data from the CLI.
// This can occur when the CLI sends malformed JSON or when the JSON structure
// doesn't match the expected schema.
type JSONDecodeError struct {
	Message string
	Raw     string // The raw JSON that failed to parse
	Cause   error
}

// Error returns the error message, implementing the error interface.
func (e *JSONDecodeError) Error() string {
	msg := e.Message
	if e.Raw != "" {
		// Truncate raw data if too long
		rawSnippet := e.Raw
		if len(rawSnippet) > 100 {
			rawSnippet = rawSnippet[:100] + "..."
		}
		msg = fmt.Sprintf("%s (raw: %s)", msg, rawSnippet)
	}
	if e.Cause != nil {
		msg = msg + ": " + e.Cause.Error()
	}
	return msg
}

// Is checks if the target error is a JSONDecodeError.
func (e *JSONDecodeError) Is(target error) bool {
	_, ok := target.(*JSONDecodeError)
	return ok
}

// Unwrap returns the wrapped error.
func (e *JSONDecodeError) Unwrap() error {
	return e.Cause
}

// NewJSONDecodeError creates a new JSONDecodeError with the given message.
func NewJSONDecodeError(message string) *JSONDecodeError {
	return &JSONDecodeError{Message: message}
}

// NewJSONDecodeErrorWithRaw creates a new JSONDecodeError with the given message and raw JSON.
func NewJSONDecodeErrorWithRaw(message string, raw string) *JSONDecodeError {
	return &JSONDecodeError{Message: message, Raw: raw}
}

// NewJSONDecodeErrorWithCause creates a new JSONDecodeError with the given message, raw JSON, and cause.
func NewJSONDecodeErrorWithCause(message string, raw string, cause error) *JSONDecodeError {
	return &JSONDecodeError{Message: message, Raw: raw, Cause: cause}
}

// MessageParseError indicates a failure to parse a message from the CLI.
// This differs from JSONDecodeError in that the JSON may be valid but the message
// structure is invalid or unexpected (e.g., missing required fields, wrong types).
type MessageParseError struct {
	Message     string
	MessageType string // The type of message that failed to parse
	Cause       error
}

// Error returns the error message, implementing the error interface.
func (e *MessageParseError) Error() string {
	msg := e.Message
	if e.MessageType != "" {
		msg = fmt.Sprintf("%s (type: %s)", msg, e.MessageType)
	}
	if e.Cause != nil {
		msg = msg + ": " + e.Cause.Error()
	}
	return msg
}

// Is checks if the target error is a MessageParseError.
func (e *MessageParseError) Is(target error) bool {
	_, ok := target.(*MessageParseError)
	return ok
}

// Unwrap returns the wrapped error.
func (e *MessageParseError) Unwrap() error {
	return e.Cause
}

// NewMessageParseError creates a new MessageParseError with the given message.
func NewMessageParseError(message string) *MessageParseError {
	return &MessageParseError{Message: message}
}

// NewMessageParseErrorWithType creates a new MessageParseError with the given message and message type.
func NewMessageParseErrorWithType(message string, messageType string) *MessageParseError {
	return &MessageParseError{Message: message, MessageType: messageType}
}

// NewMessageParseErrorWithCause creates a new MessageParseError with the given message, message type, and cause.
func NewMessageParseErrorWithCause(message string, messageType string, cause error) *MessageParseError {
	return &MessageParseError{Message: message, MessageType: messageType, Cause: cause}
}

// ControlProtocolError indicates a violation of the control protocol between
// the SDK and CLI. This includes invalid request/response sequences, unexpected
// control messages, or protocol version mismatches.
type ControlProtocolError struct {
	Message string
	Cause   error
}

// Error returns the error message, implementing the error interface.
func (e *ControlProtocolError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

// Is checks if the target error is a ControlProtocolError.
func (e *ControlProtocolError) Is(target error) bool {
	_, ok := target.(*ControlProtocolError)
	return ok
}

// Unwrap returns the wrapped error.
func (e *ControlProtocolError) Unwrap() error {
	return e.Cause
}

// NewControlProtocolError creates a new ControlProtocolError with the given message.
func NewControlProtocolError(message string) *ControlProtocolError {
	return &ControlProtocolError{Message: message}
}

// NewControlProtocolErrorWithCause creates a new ControlProtocolError with the given message and cause.
func NewControlProtocolErrorWithCause(message string, cause error) *ControlProtocolError {
	return &ControlProtocolError{Message: message, Cause: cause}
}

// PermissionDeniedError indicates that a permission request was denied.
// This occurs when the user or permission callback denies a tool use request,
// or when a permission check fails.
type PermissionDeniedError struct {
	Message  string
	ToolName string // The tool that was denied
	Reason   string // Optional reason for denial
	Cause    error
}

// Error returns the error message, implementing the error interface.
func (e *PermissionDeniedError) Error() string {
	msg := e.Message
	if e.ToolName != "" {
		msg = fmt.Sprintf("%s (tool: %s)", msg, e.ToolName)
	}
	if e.Reason != "" {
		msg = fmt.Sprintf("%s - %s", msg, e.Reason)
	}
	if e.Cause != nil {
		msg = msg + ": " + e.Cause.Error()
	}
	return msg
}

// Is checks if the target error is a PermissionDeniedError.
func (e *PermissionDeniedError) Is(target error) bool {
	_, ok := target.(*PermissionDeniedError)
	return ok
}

// Unwrap returns the wrapped error.
func (e *PermissionDeniedError) Unwrap() error {
	return e.Cause
}

// NewPermissionDeniedError creates a new PermissionDeniedError with the given message.
func NewPermissionDeniedError(message string) *PermissionDeniedError {
	return &PermissionDeniedError{Message: message}
}

// NewPermissionDeniedErrorWithTool creates a new PermissionDeniedError with the given message and tool name.
func NewPermissionDeniedErrorWithTool(message string, toolName string) *PermissionDeniedError {
	return &PermissionDeniedError{Message: message, ToolName: toolName}
}

// NewPermissionDeniedErrorWithReason creates a new PermissionDeniedError with the given message, tool name, and reason.
func NewPermissionDeniedErrorWithReason(message string, toolName string, reason string) *PermissionDeniedError {
	return &PermissionDeniedError{Message: message, ToolName: toolName, Reason: reason}
}

// NewPermissionDeniedErrorWithCause creates a new PermissionDeniedError with the given message and cause.
func NewPermissionDeniedErrorWithCause(message string, cause error) *PermissionDeniedError {
	return &PermissionDeniedError{Message: message, Cause: cause}
}

// Helper functions for error checking

// IsCLINotFoundError checks if an error is or wraps a CLINotFoundError.
func IsCLINotFoundError(err error) bool {
	var e *CLINotFoundError
	return errors.As(err, &e)
}

// IsCLIConnectionError checks if an error is or wraps a CLIConnectionError.
func IsCLIConnectionError(err error) bool {
	var e *CLIConnectionError
	return errors.As(err, &e)
}

// IsProcessError checks if an error is or wraps a ProcessError.
func IsProcessError(err error) bool {
	var e *ProcessError
	return errors.As(err, &e)
}

// IsCLIJSONDecodeError checks if an error is or wraps a CLIJSONDecodeError.
func IsCLIJSONDecodeError(err error) bool {
	var e *CLIJSONDecodeError
	return errors.As(err, &e)
}

// IsJSONDecodeError checks if an error is or wraps a JSONDecodeError.
func IsJSONDecodeError(err error) bool {
	var e *JSONDecodeError
	return errors.As(err, &e)
}

// IsMessageParseError checks if an error is or wraps a MessageParseError.
func IsMessageParseError(err error) bool {
	var e *MessageParseError
	return errors.As(err, &e)
}

// IsControlProtocolError checks if an error is or wraps a ControlProtocolError.
func IsControlProtocolError(err error) bool {
	var e *ControlProtocolError
	return errors.As(err, &e)
}

// IsPermissionDeniedError checks if an error is or wraps a PermissionDeniedError.
func IsPermissionDeniedError(err error) bool {
	var e *PermissionDeniedError
	return errors.As(err, &e)
}

// SessionNotFoundError indicates that a Claude session could not be found.
// This typically occurs when attempting to resume a conversation with a session ID
// that no longer exists in Claude's database, often due to CLI reinstallation or
// session expiration.
type SessionNotFoundError struct {
	SessionID string // The session ID that was not found
	Message   string // Human-readable error message
	Cause     error  // Optional underlying error
}

// Error returns the error message, implementing the error interface.
func (e *SessionNotFoundError) Error() string {
	msg := e.Message
	if e.SessionID != "" {
		msg = fmt.Sprintf("%s (session ID: %s)", msg, e.SessionID)
	}
	if e.Cause != nil {
		msg = msg + ": " + e.Cause.Error()
	}
	return msg
}

// Is checks if the target error is a SessionNotFoundError.
func (e *SessionNotFoundError) Is(target error) bool {
	_, ok := target.(*SessionNotFoundError)
	return ok
}

// Unwrap returns the wrapped error.
func (e *SessionNotFoundError) Unwrap() error {
	return e.Cause
}

// NewSessionNotFoundError creates a new SessionNotFoundError with the given session ID and message.
func NewSessionNotFoundError(sessionID, message string) *SessionNotFoundError {
	return &SessionNotFoundError{
		SessionID: sessionID,
		Message:   message,
	}
}

// NewSessionNotFoundErrorWithCause creates a new SessionNotFoundError with the given session ID, message, and cause.
func NewSessionNotFoundErrorWithCause(sessionID, message string, cause error) *SessionNotFoundError {
	return &SessionNotFoundError{
		SessionID: sessionID,
		Message:   message,
		Cause:     cause,
	}
}

// IsSessionNotFoundError checks if an error is or wraps a SessionNotFoundError.
func IsSessionNotFoundError(err error) bool {
	var e *SessionNotFoundError
	return errors.As(err, &e)
}

// WithContext wraps an error with additional context information.
// This helps in debugging by providing more information about where the error occurred.
func WithContext(err error, context string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", context, err)
}

// ErrorWithContext represents an error with additional context information.
type ErrorWithContext struct {
	Err      error
	Context  string
	Location string
}

// Error implements the error interface.
func (e *ErrorWithContext) Error() string {
	if e.Context != "" {
		return fmt.Sprintf("[%s] %s: %v", e.Location, e.Context, e.Err)
	}
	return fmt.Sprintf("[%s] %v", e.Location, e.Err)
}

// Unwrap returns the underlying error.
func (e *ErrorWithContext) Unwrap() error {
	return e.Err
}

// Is checks if target matches this error or its wrapped error.
func (e *ErrorWithContext) Is(target error) bool {
	return target == e || errors.Is(e.Err, target)
}

// NewErrorWithContext creates a new ErrorWithContext with the given error, context, and location.
func NewErrorWithContext(err error, context, location string) *ErrorWithContext {
	return &ErrorWithContext{
		Err:      err,
		Context:  context,
		Location: location,
	}
}
