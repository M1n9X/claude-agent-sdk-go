package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/M1n9X/claude-agent-sdk-go/internal/log"
	"github.com/M1n9X/claude-agent-sdk-go/types"
)

const (
	// SDKVersion is the version identifier for this SDK
	SDKVersion = "0.1.0"
)

// SubprocessCLITransport implements Transport using a Claude Code CLI subprocess.
// It manages the subprocess lifecycle, stdin/stdout/stderr pipes, and message streaming.
type SubprocessCLITransport struct {
	cliPath         string
	cwd             string
	env             map[string]string
	logger          *log.Logger
	resumeSessionID string                    // Optional session ID to resume conversation
	options         *types.ClaudeAgentOptions // Options for CLI configuration

	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	stderr io.ReadCloser

	ctx    context.Context
	cancel context.CancelFunc

	// Message streaming
	messages chan types.Message

	// Writer for stdin
	writer *JSONLineWriter

	// MCP configuration file paths (will be cleaned up on Close)
	mcpConfigFiles []string

	// Error tracking
	mu    sync.Mutex
	err   error
	ready bool
}

// NewSubprocessCLITransport creates a new transport instance.
// The cliPath should point to the claude binary.
// The cwd is the working directory for the subprocess (empty string uses current directory).
// The env map contains additional environment variables to set for the subprocess.
// The logger is used for debug/diagnostic output.
// The resumeSessionID is an optional session ID to resume a previous conversation.
// The options contains configuration for the CLI.
func NewSubprocessCLITransport(cliPath, cwd string, env map[string]string, logger *log.Logger, resumeSessionID string, options *types.ClaudeAgentOptions) *SubprocessCLITransport {
	// Use configurable message channel capacity with default of 10
	capacity := 10
	if options != nil && options.MessageChannelCapacity != nil {
		capacity = *options.MessageChannelCapacity
	}

	return &SubprocessCLITransport{
		cliPath:         cliPath,
		cwd:             cwd,
		env:             env,
		logger:          logger,
		resumeSessionID: resumeSessionID,
		options:         options,
		messages:        make(chan types.Message, capacity), // Buffered channel for smooth streaming
	}
}

// Connect starts the Claude Code CLI subprocess and establishes communication pipes.
// It launches the subprocess with "agent --stdio" arguments and sets up the environment.
func (t *SubprocessCLITransport) Connect(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.cmd != nil {
		return nil // Already connected
	}

	t.logger.Debug("Starting Claude CLI subprocess: %s", t.cliPath)

	// Create cancellable context
	t.ctx, t.cancel = context.WithCancel(ctx)

	// Build command arguments
	args := t.buildCommandArgs()

	// Log the full command for debugging
	t.logger.Debug("Claude CLI command: %s %v", t.cliPath, args)

	// Create command with arguments
	t.cmd = exec.CommandContext(t.ctx, t.cliPath, args...)

	// Set working directory if provided
	if t.cwd != "" {
		t.cmd.Dir = t.cwd
	}

	// Set up environment variables
	// Start with current environment
	t.cmd.Env = os.Environ()

	// Add SDK-specific variables
	t.cmd.Env = append(t.cmd.Env, "CLAUDE_CODE_ENTRYPOINT=agent")
	t.cmd.Env = append(t.cmd.Env, fmt.Sprintf("CLAUDE_AGENT_SDK_VERSION=%s", SDKVersion))

	// Add model environment variable if specified in options (ANTHROPIC_MODEL)
	// This is critical - both CLI flag and env var should be set for maximum compatibility
	if t.options != nil && t.options.Model != nil {
		t.cmd.Env = append(t.cmd.Env, fmt.Sprintf("ANTHROPIC_MODEL=%s", *t.options.Model))
		t.logger.Debug("Setting ANTHROPIC_MODEL environment variable: %s", *t.options.Model)
	} else {
		t.logger.Debug("ANTHROPIC_MODEL not set (using CLI default)")
	}

	// Add base URL environment variable if specified in options (ANTHROPIC_BASE_URL)
	// If not set, Claude CLI will use default Anthropic API endpoint
	if t.options != nil && t.options.BaseURL != nil {
		t.cmd.Env = append(t.cmd.Env, fmt.Sprintf("ANTHROPIC_BASE_URL=%s", *t.options.BaseURL))
		t.logger.Debug("Setting ANTHROPIC_BASE_URL environment variable: %s", *t.options.BaseURL)
	} else {
		t.logger.Debug("ANTHROPIC_BASE_URL not set (using default Anthropic API)")
	}

	// Add custom environment variables (these can override the above if needed)
	for key, value := range t.env {
		t.cmd.Env = append(t.cmd.Env, fmt.Sprintf("%s=%s", key, value))
		t.logger.Debug("Setting custom environment variable: %s=%s", key, value)
	}

	// Set up pipes
	var err error

	t.stdin, err = t.cmd.StdinPipe()
	if err != nil {
		return types.NewCLIConnectionErrorWithCause("failed to create stdin pipe", err)
	}

	t.stdout, err = t.cmd.StdoutPipe()
	if err != nil {
		return types.NewCLIConnectionErrorWithCause("failed to create stdout pipe", err)
	}

	t.stderr, err = t.cmd.StderrPipe()
	if err != nil {
		return types.NewCLIConnectionErrorWithCause("failed to create stderr pipe", err)
	}

	// Start the process
	if err := t.cmd.Start(); err != nil {
		t.logger.Error("Failed to start subprocess: %v", err)
		return types.NewCLIConnectionErrorWithCause("failed to start subprocess", err)
	}
	t.logger.Debug("CLI subprocess started successfully (PID: %d)", t.cmd.Process.Pid)

	// Create JSON line writer for stdin
	t.writer = NewJSONLineWriter(t.stdin)

	// Launch message reader loop in goroutine
	go t.messageReaderLoop(t.ctx)

	// Launch stderr reader for debugging
	go t.readStderr(t.ctx)

	// Mark as ready
	t.ready = true
	t.logger.Debug("Transport ready for communication")

	return nil
}

// messageReaderLoop reads JSON lines from stdout and parses them into messages.
// It runs in a goroutine and sends messages to the messages channel.
// It respects context cancellation and closes the messages channel when done.
func (t *SubprocessCLITransport) messageReaderLoop(ctx context.Context) {
	defer close(t.messages)

	t.logger.Debug("Message reader loop started")
	reader := NewJSONLineReader(t.stdout)

	for {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			t.logger.Debug("Message reader loop stopped: context cancelled")
			return
		default:
		}

		// Read next JSON line
		line, err := reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				t.logger.Debug("Message reader loop stopped: EOF from CLI")
				// Normal end of stream
				return
			}

			t.logger.Error("Failed to read from CLI stdout: %v", err)
			// Store error and return
			t.OnError(types.NewJSONDecodeErrorWithCause(
				"failed to read JSON line from subprocess",
				string(line),
				err,
			))
			return
		}

		// Skip empty lines
		if len(line) == 0 {
			continue
		}

		// Parse JSON into message
		msg, err := types.UnmarshalMessage(line)
		if err != nil {
			t.logger.Warning("Failed to parse message from CLI: %v", err)
			// Store parse error but continue reading
			t.OnError(err)
			continue
		}

		t.logger.Debug("Received message from CLI: type=%s", msg.GetMessageType())

		// Send message to channel (respect context cancellation)
		select {
		case <-ctx.Done():
			return
		case t.messages <- msg:
			// Message sent successfully
		}
	}
}

// Write sends a JSON message to the subprocess stdin.
// The data should be a complete JSON string (newline will be added automatically).
func (t *SubprocessCLITransport) Write(ctx context.Context, data string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.ready {
		return types.NewCLIConnectionError("transport is not ready for writing")
	}

	if t.writer == nil {
		return types.NewCLIConnectionError("stdin writer not initialized")
	}

	t.logger.Debug("Sending message to CLI stdin")

	// Write JSON line (includes newline and flush)
	if err := t.writer.WriteLine(data); err != nil {
		t.ready = false
		t.err = types.NewCLIConnectionErrorWithCause("failed to write to subprocess stdin", err)
		t.logger.Error("Failed to write to CLI stdin: %v", err)
		return t.err
	}

	return nil
}

// ReadMessages returns a channel of incoming messages from the subprocess.
// The channel is closed when the subprocess exits or an error occurs.
func (t *SubprocessCLITransport) ReadMessages(ctx context.Context) <-chan types.Message {
	return t.messages
}

// buildCommandArgs builds the command line arguments for the CLI subprocess.
// This is extracted into a separate method to allow for testing.
func (t *SubprocessCLITransport) buildCommandArgs() []string {
	args := []string{
		"--input-format=stream-json",
		"--output-format=stream-json",
		"--verbose",
	}

	opts := t.options

	// Add permission prompt tool if specified
	if opts != nil && opts.PermissionPromptToolName != nil {
		args = append(args, "--permission-prompt-tool", *opts.PermissionPromptToolName)
		t.logger.Debug("Setting permission prompt tool: %s", *opts.PermissionPromptToolName)
	}

	// Add permission mode if specified
	if opts != nil && opts.PermissionMode != nil {
		args = append(args, "--permission-mode", string(*opts.PermissionMode))
		t.logger.Debug("Setting permission mode: %s", string(*opts.PermissionMode))
	}

	// Configure system prompt
	// By default, use an empty system prompt to avoid implicit presets from the CLI.
	if opts == nil || opts.SystemPrompt == nil {
		args = append(args, "--system-prompt", "")
		t.logger.Debug("Using empty system prompt (no preset)")
	} else {
		switch prompt := opts.SystemPrompt.(type) {
		case string:
			args = append(args, "--system-prompt", prompt)
			t.logger.Debug("Setting system prompt: %s", prompt)
		case types.SystemPromptPreset:
			if prompt.Type == "preset" && prompt.Append != nil {
				args = append(args, "--append-system-prompt", *prompt.Append)
				t.logger.Debug("Appending to preset system prompt")
			}
		case *types.SystemPromptPreset:
			if prompt != nil && prompt.Type == "preset" && prompt.Append != nil {
				args = append(args, "--append-system-prompt", *prompt.Append)
				t.logger.Debug("Appending to preset system prompt")
			}
		}
	}

	// Add model if specified
	if opts != nil && opts.Model != nil {
		args = append(args, "--model", *opts.Model)
		t.logger.Debug("Setting model: %s", *opts.Model)
	}

	// Add fallback model if specified
	if opts != nil && opts.FallbackModel != nil {
		args = append(args, "--fallback-model", *opts.FallbackModel)
		t.logger.Debug("Setting fallback model: %s", *opts.FallbackModel)
	}

	// Add --resume flag if resuming a conversation
	if t.resumeSessionID != "" {
		args = append(args, "--resume", t.resumeSessionID)
		t.logger.Debug("Resuming Claude CLI conversation with session ID: %s", t.resumeSessionID)
	}

	// Add --fork-session flag if forking a resumed session
	if t.options != nil && t.options.ForkSession {
		args = append(args, "--fork-session")
		t.logger.Debug("Forking resumed session to new session ID")
	}

	// Add permission bypass flags if enabled
	if opts != nil {
		// Must set allow flag first (acts as safety switch)
		if opts.AllowDangerouslySkipPermissions {
			args = append(args, "--allow-dangerously-skip-permissions")
			t.logger.Debug("Allowing permission bypass (safety switch enabled)")

			// Only add skip flag if allow flag is also set
			if opts.DangerouslySkipPermissions {
				args = append(args, "--dangerously-skip-permissions")
				t.logger.Debug("DANGER: Bypassing all permissions - use only in sandboxed environments!")
			}
		}
	}

	// Allowed and disallowed tools
	if opts != nil && len(opts.AllowedTools) > 0 {
		args = append(args, "--allowedTools", strings.Join(opts.AllowedTools, ","))
		t.logger.Debug("Setting allowed tools: %v", opts.AllowedTools)
	}

	if opts != nil && len(opts.DisallowedTools) > 0 {
		args = append(args, "--disallowedTools", strings.Join(opts.DisallowedTools, ","))
		t.logger.Debug("Setting disallowed tools: %v", opts.DisallowedTools)
	}

	// Conversation controls
	if opts != nil && opts.ContinueConversation {
		args = append(args, "--continue")
		t.logger.Debug("Continuing previous conversation")
	}

	if opts != nil && opts.MaxTurns != nil {
		args = append(args, "--max-turns", strconv.Itoa(*opts.MaxTurns))
		t.logger.Debug("Setting max turns: %d", *opts.MaxTurns)
	}

	// Add MCP servers configuration
	if opts != nil && opts.McpServers != nil {
		switch servers := opts.McpServers.(type) {
		case map[string]interface{}:
			if len(servers) > 0 {
				if configFile := t.generateMcpConfigFile(); configFile != "" {
					args = append(args, "--mcp-servers", configFile)
					t.logger.Debug("Setting MCP servers config: %s", configFile)
				}
			}
		case string:
			if strings.TrimSpace(servers) != "" {
				args = append(args, "--mcp-servers", servers)
				t.logger.Debug("Using MCP servers config path: %s", servers)
			}
		case *string:
			if servers != nil && strings.TrimSpace(*servers) != "" {
				args = append(args, "--mcp-servers", *servers)
				t.logger.Debug("Using MCP servers config path: %s", *servers)
			}
		}
	}

	// Add extended thinking token limit if specified
	if opts != nil && opts.MaxThinkingTokens != nil {
		args = append(args, "--max-thinking-tokens", fmt.Sprintf("%d", *opts.MaxThinkingTokens))
		t.logger.Debug("Setting max thinking tokens: %d", *opts.MaxThinkingTokens)
	}

	// Add budget limit if specified
	if opts != nil && opts.MaxBudgetUSD != nil {
		args = append(args, "--max-budget-usd", fmt.Sprintf("%.2f", *opts.MaxBudgetUSD))
		t.logger.Debug("Setting max budget: $%.2f USD", *opts.MaxBudgetUSD)
	}

	// Settings and directories
	if opts != nil && opts.Settings != nil {
		args = append(args, "--settings", *opts.Settings)
		t.logger.Debug("Setting settings path: %s", *opts.Settings)
	}

	if opts != nil && len(opts.AddDirs) > 0 {
		for _, dir := range opts.AddDirs {
			args = append(args, "--add-dir", dir)
			t.logger.Debug("Adding directory: %s", dir)
		}
	}

	if opts != nil && opts.SettingSources != nil {
		sources := make([]string, len(opts.SettingSources))
		for i, s := range opts.SettingSources {
			sources[i] = string(s)
		}
		sourcesValue := strings.Join(sources, ",")
		args = append(args, "--setting-sources", sourcesValue)
		t.logger.Debug("Setting sources: %s", sourcesValue)
	}

	// Include partial messages
	if opts != nil && opts.IncludePartialMessages {
		args = append(args, "--include-partial-messages")
		t.logger.Debug("Including partial messages")
	}

	// Agent definitions
	if opts != nil && len(opts.Agents) > 0 {
		agentsPayload := make(map[string]map[string]interface{}, len(opts.Agents))
		for name, agent := range opts.Agents {
			agentConfig := map[string]interface{}{
				"description": agent.Description,
				"prompt":      agent.Prompt,
			}
			if len(agent.Tools) > 0 {
				agentConfig["tools"] = agent.Tools
			}
			if agent.Model != nil {
				agentConfig["model"] = *agent.Model
			}
			agentsPayload[name] = agentConfig
		}

		if data, err := json.Marshal(agentsPayload); err == nil {
			args = append(args, "--agents", string(data))
			t.logger.Debug("Setting agents configuration")
		} else {
			t.logger.Warning("Failed to marshal agents configuration: %v", err)
		}
	}

	// Plugins
	if opts != nil && len(opts.Plugins) > 0 {
		for _, plugin := range opts.Plugins {
			pluginType := plugin.Type
			if pluginType == "" {
				pluginType = "local"
			}
			switch pluginType {
			case "local":
				args = append(args, "--plugin-dir", plugin.Path)
				t.logger.Debug("Adding plugin directory: %s", plugin.Path)
			default:
				t.logger.Warning("Unsupported plugin type %q, skipping", pluginType)
			}
		}
	}

	// Extra args
	if opts != nil && len(opts.ExtraArgs) > 0 {
		for flag, value := range opts.ExtraArgs {
			if value == nil {
				args = append(args, fmt.Sprintf("--%s", flag))
			} else {
				args = append(args, fmt.Sprintf("--%s", flag), *value)
			}
		}
	}

	return args
}

// generateMcpConfigFile generates a temporary MCP configuration file for external MCP servers.
// SDK MCP servers are handled in-process and do not need a config file.
func (t *SubprocessCLITransport) generateMcpConfigFile() string {
	if t.options == nil || t.options.McpServers == nil {
		return ""
	}

	// Type assert to the expected map type
	servers, ok := t.options.McpServers.(map[string]interface{})
	if !ok {
		return ""
	}

	if len(servers) == 0 {
		return ""
	}

	config := map[string]interface{}{
		"mcpServers": make(map[string]interface{}),
	}

	externalServers := false

	for name, server := range servers {
		// Skip SDK MCP servers (handled in-process)
		if _, ok := server.(*types.ToolServerConfig); ok {
			continue
		}

		// Add to config based on type
		switch s := server.(type) {
		case types.McpStdioServerConfig:
			serverConfig := map[string]interface{}{
				"type":    "stdio",
				"command": s.Command,
				"args":    s.Args,
			}
			if len(s.Env) > 0 {
				serverConfig["env"] = s.Env
			}
			config["mcpServers"].(map[string]interface{})[name] = serverConfig
			externalServers = true

		case types.McpSSEServerConfig:
			config["mcpServers"].(map[string]interface{})[name] = map[string]interface{}{
				"type":    "sse",
				"url":     s.URL,
				"headers": s.Headers,
			}
			externalServers = true

		case types.McpHTTPServerConfig:
			config["mcpServers"].(map[string]interface{})[name] = map[string]interface{}{
				"type":    "http",
				"url":     s.URL,
				"headers": s.Headers,
			}
			externalServers = true
		}
	}

	// If no external servers, return empty string
	if !externalServers {
		return ""
	}

	// Create temp file
	tmpFile, err := os.CreateTemp("", "claude-mcp-*.json")
	if err != nil {
		t.logger.Error("Failed to create MCP config file: %v", err)
		return ""
	}
	defer tmpFile.Close()

	encoder := json.NewEncoder(tmpFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(config); err != nil {
		t.logger.Error("Failed to write MCP config file: %v", err)
		os.Remove(tmpFile.Name())
		return ""
	}

	// Store path for cleanup
	t.mcpConfigFiles = append(t.mcpConfigFiles, tmpFile.Name())
	return tmpFile.Name()
}

// Close terminates the subprocess and cleans up all resources.
// It attempts to gracefully shut down the subprocess with a timeout.
func (t *SubprocessCLITransport) Close(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.cmd == nil {
		return nil // Not connected
	}

	t.logger.Debug("Closing CLI subprocess...")
	t.ready = false

	// Cancel the context to stop goroutines
	if t.cancel != nil {
		t.cancel()
		t.cancel = nil
	}

	// Close stdin to signal end of input
	if t.stdin != nil {
		_ = t.stdin.Close()
		t.stdin = nil
	}

	// Cleanup MCP config files if they exist
	for _, configFile := range t.mcpConfigFiles {
		os.Remove(configFile)
	}
	t.mcpConfigFiles = nil

	// Wait for process to exit (with context timeout)
	done := make(chan error, 1)
	go func() {
		done <- t.cmd.Wait()
	}()

	select {
	case <-ctx.Done():
		// Timeout - kill the process
		if t.cmd.Process != nil {
			_ = t.cmd.Process.Kill()
		}
		<-done // Wait for Wait() to return
		return types.NewProcessError("subprocess did not exit gracefully, killed")

	case err := <-done:
		// Process exited
		// During normal shutdown, the subprocess may exit with non-zero codes
		// which is expected behavior when stdin is closed, so we don't treat
		// these as errors during the Close operation
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				// Log the exit code for debugging but don't return it as an error
				// since this is expected during normal shutdown
				t.logger.Debug("CLI subprocess exited with code %d during shutdown (expected)", exitErr.ExitCode())
				return nil
			}
			// For other types of errors, log but don't return as error during shutdown
			t.logger.Debug("CLI subprocess error during shutdown: %v", err)
			return nil
		}
		return nil
	}
}

// OnError stores an error that occurred during transport operation.
// This allows errors from the reading loop to be retrieved later.
func (t *SubprocessCLITransport) OnError(err error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.err == nil {
		t.err = err
	}
}

// IsReady returns true if the transport is ready for communication.
func (t *SubprocessCLITransport) IsReady() bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.ready
}

// GetError returns any error that occurred during transport operation.
// This is useful for checking if an error occurred in the reading loop.
func (t *SubprocessCLITransport) GetError() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	return t.err
}

// readStderr reads stderr output in a goroutine for debugging.
// This is a helper function for monitoring subprocess errors.
// It also parses known error patterns and stores them as typed errors.
func (t *SubprocessCLITransport) readStderr(ctx context.Context) {
	if t.stderr == nil {
		return
	}

	// Open log file for stderr output
	homeDir, _ := os.UserHomeDir()
	logPath := fmt.Sprintf("%s/.claude/agents_server/cli_stderr.log", homeDir)

	// Create the directory if it doesn't exist
	logDir := filepath.Dir(logPath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		// Fallback to stderr if directory can't be created
		fmt.Fprintf(os.Stderr, "[SDK] Failed to create log directory: %v\n", err)
		return
	}

	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		// Fallback to stderr if file can't be opened
		fmt.Fprintf(os.Stderr, "[SDK] Failed to open stderr log file: %v\n", err)
		return
	}
	defer func() {
		_ = logFile.Close()
	}()

	reader := NewJSONLineReader(t.stderr)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		line, err := reader.ReadLine()
		if err != nil {
			return
		}

		// Log stderr output to file
		if len(line) > 0 {
			stderrText := string(line)
			_, _ = fmt.Fprintf(logFile, "[Claude CLI stderr]: %s\n", stderrText)
			_ = logFile.Sync() // Flush to disk immediately

			// Parse known error patterns and create typed errors
			t.parseStderrError(stderrText)
		}
	}
}

// parseStderrError parses stderr text for known error patterns and stores typed errors.
func (t *SubprocessCLITransport) parseStderrError(stderrText string) {
	// Check for "No conversation found with session ID:" error
	if matched, sessionID := extractSessionNotFoundError(stderrText); matched {
		// Create typed error
		err := types.NewSessionNotFoundError(
			sessionID,
			"Claude CLI could not find this conversation. It may have been deleted or the CLI was reinstalled.",
		)

		// Store error for retrieval
		t.OnError(err)

		// Log it
		t.logger.Error("Claude session not found: %s", sessionID)
	}
}

// extractSessionNotFoundError checks if the stderr text contains a session not found error.
// Returns (true, sessionID) if matched, (false, "") otherwise.
func extractSessionNotFoundError(stderrText string) (bool, string) {
	// Pattern: "No conversation found with session ID: <uuid>"
	// Example: "No conversation found with session ID: 8587b432-e504-42c8-b9a7-e3fd0b4b2c60"
	const pattern = "No conversation found with session ID:"

	if idx := findSubstring(stderrText, pattern); idx >= 0 {
		// Extract session ID after the pattern
		sessionIDStart := idx + len(pattern)
		if sessionIDStart < len(stderrText) {
			// Trim whitespace and extract the session ID
			remaining := stderrText[sessionIDStart:]
			sessionID := trimWhitespace(remaining)
			// Session ID is the first token (UUID format)
			if len(sessionID) > 0 {
				// Take everything up to the first whitespace or end of string
				endIdx := 0
				for endIdx < len(sessionID) && !isWhitespace(rune(sessionID[endIdx])) {
					endIdx++
				}
				sessionID = sessionID[:endIdx]
				return true, sessionID
			}
		}
	}

	return false, ""
}

// Helper functions for string parsing

func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func trimWhitespace(s string) string {
	start := 0
	for start < len(s) && isWhitespace(rune(s[start])) {
		start++
	}
	end := len(s)
	for end > start && isWhitespace(rune(s[end-1])) {
		end--
	}
	return s[start:end]
}

func isWhitespace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r'
}
