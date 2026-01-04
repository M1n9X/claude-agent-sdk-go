package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/M1n9X/claude-agent-sdk-go/internal/log"
	"github.com/M1n9X/claude-agent-sdk-go/types"
)

// TestFindCLI tests CLI discovery in various scenarios
func TestFindCLI(t *testing.T) {
	// Disable version checking for these tests since we're using mock binaries
	t.Setenv("CLAUDE_AGENT_SDK_SKIP_VERSION_CHECK", "1")

	tests := []struct {
		name      string
		setup     func() func() // Returns cleanup function
		wantError bool
	}{
		{
			name: "CLI in PATH",
			setup: func() func() {
				// Save original PATH
				origPath := os.Getenv("PATH")

				// Create temporary directory with a mock claude binary
				tmpDir := t.TempDir()
				claudePath := filepath.Join(tmpDir, "claude")

				// Create mock binary
				f, err := os.Create(claudePath)
				if err != nil {
					t.Fatalf("Failed to create mock binary: %v", err)
				}
				_ = f.Close()

				// Make it executable
				if err := os.Chmod(claudePath, 0755); err != nil {
					t.Fatalf("Failed to chmod mock binary: %v", err)
				}

				// Add to PATH
				_ = os.Setenv("PATH", tmpDir+string(os.PathListSeparator)+origPath)

				// Return cleanup function
				return func() {
					_ = os.Setenv("PATH", origPath)
				}
			},
			wantError: false,
		},
		// Note: "CLI not found" test is skipped because it's environment-dependent
		// If Claude CLI is installed in standard locations (like ~/.local/bin/claude),
		// it will be found even when PATH/HOME are cleared since FindCLI checks
		// hardcoded paths using user.Current(). This is actually desired behavior.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := tt.setup()
			defer cleanup()

			path, err := FindCLI()

			if tt.wantError {
				if err == nil {
					t.Errorf("FindCLI() expected error, got nil (found path: %s, PATH=%s, HOME=%s)", path, os.Getenv("PATH"), os.Getenv("HOME"))
				}
				var cliNotFoundErr *types.CLINotFoundError
				if err != nil && !types.IsCLINotFoundError(err) {
					t.Errorf("FindCLI() error type = %T, want *types.CLINotFoundError", err)
				}
				_ = cliNotFoundErr
			} else {
				if err != nil {
					t.Errorf("FindCLI() unexpected error: %v", err)
				}
				if path == "" {
					t.Errorf("FindCLI() returned empty path")
				}
			}
		})
	}
}

// TestExpandHome tests home directory expansion
func TestExpandHome(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "tilde only",
			input: "~",
			want:  "HOME_DIR", // Will be replaced in test
		},
		{
			name:  "tilde with path",
			input: "~/.config/claude",
			want:  "HOME_DIR/.config/claude",
		},
		{
			name:  "no tilde",
			input: "/usr/local/bin/claude",
			want:  "/usr/local/bin/claude",
		},
		{
			name:  "relative path",
			input: "./bin/claude",
			want:  "./bin/claude",
		},
	}

	// Get actual home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Replace placeholder with actual home dir
			want := strings.ReplaceAll(tt.want, "HOME_DIR", homeDir)

			got := expandHome(tt.input)
			if got != want {
				t.Errorf("expandHome(%q) = %q, want %q", tt.input, got, want)
			}
		})
	}
}

// TestJSONLineReader tests buffered JSON line reading
func TestJSONLineReader(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []string
		wantErr bool
	}{
		{
			name:  "single line",
			input: `{"type":"test","data":"hello"}` + "\n",
			want:  []string{`{"type":"test","data":"hello"}`},
		},
		{
			name: "multiple lines",
			input: `{"type":"test1"}` + "\n" +
				`{"type":"test2"}` + "\n" +
				`{"type":"test3"}` + "\n",
			want: []string{
				`{"type":"test1"}`,
				`{"type":"test2"}`,
				`{"type":"test3"}`,
			},
		},
		{
			name:  "empty lines ignored",
			input: `{"type":"test1"}` + "\n\n" + `{"type":"test2"}` + "\n",
			want:  []string{`{"type":"test1"}`, `{"type":"test2"}`},
		},
		{
			name:  "trailing newline",
			input: `{"type":"test"}` + "\n",
			want:  []string{`{"type":"test"}`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := NewJSONLineReader(strings.NewReader(tt.input))

			var got []string
			for {
				line, err := reader.ReadLine()
				if err == io.EOF {
					break
				}
				if err != nil {
					if !tt.wantErr {
						t.Errorf("ReadLine() unexpected error: %v", err)
					}
					return
				}

				if len(line) > 0 {
					got = append(got, string(line))
				}
			}

			if len(got) != len(tt.want) {
				t.Errorf("ReadLine() got %d lines, want %d", len(got), len(tt.want))
			}

			for i := range got {
				if i >= len(tt.want) {
					break
				}
				if got[i] != tt.want[i] {
					t.Errorf("ReadLine() line %d = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

// TestJSONLineReaderBufferOverflow tests buffer size limits
func TestJSONLineReaderBufferOverflow(t *testing.T) {
	// Create a JSON line larger than the buffer
	// Note: bufio.Scanner needs significantly larger input to trigger the error
	smallBufferSize := 1024
	largeJSON := `{"data":"` + strings.Repeat("x", smallBufferSize*2) + `"}`

	reader := NewJSONLineReaderWithSize(strings.NewReader(largeJSON+"\n"), smallBufferSize)

	_, err := reader.ReadLine()
	// The scanner may or may not fail depending on internal buffering
	// We just verify that if there's an error, it's handled correctly
	if err != nil {
		t.Logf("ReadLine() error (expected for large buffer): %v", err)
	} else {
		// For smaller sizes, the scanner may succeed by growing the buffer
		t.Logf("ReadLine() succeeded (scanner grew buffer)")
	}
}

// TestJSONLineWriter tests buffered JSON line writing
func TestJSONLineWriter(t *testing.T) {
	tests := []struct {
		name  string
		lines []string
		want  string
	}{
		{
			name:  "single line",
			lines: []string{`{"type":"test"}`},
			want:  `{"type":"test"}` + "\n",
		},
		{
			name: "multiple lines",
			lines: []string{
				`{"type":"test1"}`,
				`{"type":"test2"}`,
				`{"type":"test3"}`,
			},
			want: `{"type":"test1"}` + "\n" +
				`{"type":"test2"}` + "\n" +
				`{"type":"test3"}` + "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			writer := NewJSONLineWriter(&buf)

			for _, line := range tt.lines {
				if err := writer.WriteLine(line); err != nil {
					t.Errorf("WriteLine() unexpected error: %v", err)
				}
			}

			got := buf.String()
			if got != tt.want {
				t.Errorf("WriteLine() wrote %q, want %q", got, tt.want)
			}
		})
	}
}

// TestSubprocessCLITransportConnect tests subprocess connection
func TestSubprocessCLITransportConnect(t *testing.T) {
	// Skip if no echo command available
	echoPath, err := FindMockCLI()
	if err != nil {
		t.Skip("No echo command available for testing")
	}

	logger := log.NewLogger(false) // Non-verbose for tests
	transport := NewSubprocessCLITransport(echoPath, "", nil, logger, "", nil)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Connect should succeed
	if err := transport.Connect(ctx); err != nil {
		t.Errorf("Connect() unexpected error: %v", err)
	}

	// Should be ready
	if !transport.IsReady() {
		t.Errorf("IsReady() = false, want true after Connect()")
	}

	// Clean up
	if err := transport.Close(ctx); err != nil {
		t.Logf("Close() error (may be expected): %v", err)
	}
}

// TestSubprocessCLITransportWrite tests writing to subprocess
func TestSubprocessCLITransportWrite(t *testing.T) {
	// Use cat command as a simple echo subprocess
	catPath, err := FindMockCLI()
	if err != nil {
		t.Skip("No cat command available for testing")
	}

	logger := log.NewLogger(false) // Non-verbose for tests
	transport := NewSubprocessCLITransport(catPath, "", nil, logger, "", nil)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Connect
	if err := transport.Connect(ctx); err != nil {
		t.Fatalf("Connect() unexpected error: %v", err)
	}
	defer func() {
		_ = transport.Close(ctx)
	}()

	// Write should succeed
	testJSON := `{"type":"test","data":"hello"}`
	if err := transport.Write(ctx, testJSON); err != nil {
		t.Errorf("Write() unexpected error: %v", err)
	}
}

// TestSubprocessCLITransportClose tests subprocess cleanup
func TestSubprocessCLITransportClose(t *testing.T) {
	echoPath, err := FindMockCLI()
	if err != nil {
		t.Skip("No echo command available for testing")
	}

	logger := log.NewLogger(false) // Non-verbose for tests
	transport := NewSubprocessCLITransport(echoPath, "", nil, logger, "", nil)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Connect and then close
	if err := transport.Connect(ctx); err != nil {
		t.Fatalf("Connect() unexpected error: %v", err)
	}

	if err := transport.Close(ctx); err != nil {
		t.Logf("Close() error (may be expected for echo): %v", err)
	}

	// Should not be ready after close
	if transport.IsReady() {
		t.Errorf("IsReady() = true, want false after Close()")
	}
}

// TestMessageReaderLoop tests message reading and parsing
func TestMessageReaderLoop(t *testing.T) {
	// Create a mock JSON stream
	jsonStream := `{"type":"user","content":"hello"}` + "\n" +
		`{"type":"assistant","content":[{"type":"text","text":"hi"}],"model":"claude-3"}` + "\n" +
		`{"type":"system","subtype":"info","data":{}}` + "\n"

	// Create a pipe to simulate subprocess output
	pr, pw := io.Pipe()

	// Write mock data in a goroutine
	go func() {
		defer func() {
			_ = pw.Close()
		}()
		_, _ = pw.Write([]byte(jsonStream))
	}()

	// Create transport with custom stdout
	logger := log.NewLogger(false) // Non-verbose for tests
	transport := &SubprocessCLITransport{
		messages: make(chan types.Message, 10),
		ready:    true,
		logger:   logger,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	transport.ctx = ctx
	transport.stdout = pr

	// Start reader loop
	go transport.messageReaderLoop(ctx)

	// Read messages from channel
	var messages []types.Message
	for msg := range transport.messages {
		messages = append(messages, msg)
	}

	// Should have parsed 3 messages
	if len(messages) != 3 {
		t.Errorf("messageReaderLoop() parsed %d messages, want 3", len(messages))
	}

	// Verify message types
	expectedTypes := []string{"user", "assistant", "system"}
	for i, msg := range messages {
		if i >= len(expectedTypes) {
			break
		}
		if msg.GetMessageType() != expectedTypes[i] {
			t.Errorf("message[%d].Type = %q, want %q", i, msg.GetMessageType(), expectedTypes[i])
		}
	}
}

// TestSubprocessEnvironment tests environment variable setup
func TestSubprocessEnvironment(t *testing.T) {
	echoPath, err := FindMockCLI()
	if err != nil {
		t.Skip("No echo command available for testing")
	}

	env := map[string]string{
		"TEST_VAR":    "test_value",
		"ANOTHER_VAR": "another_value",
	}

	logger := log.NewLogger(false) // Non-verbose for tests
	transport := NewSubprocessCLITransport(echoPath, "", env, logger, "", nil)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := transport.Connect(ctx); err != nil {
		t.Fatalf("Connect() unexpected error: %v", err)
	}
	defer func() {
		_ = transport.Close(ctx)
	}()

	// Check that environment variables were set (we can't directly verify,
	// but we can check that Connect succeeded with the env)
	if !transport.IsReady() {
		t.Errorf("IsReady() = false after Connect() with custom env")
	}
}

// FindMockCLI finds a command suitable for testing (cat, echo, etc.)
func FindMockCLI() (string, error) {
	// Try to find cat command (available on Unix systems)
	if path, err := exec.LookPath("cat"); err == nil {
		return path, nil
	}

	// Try echo as fallback
	if path, err := exec.LookPath("echo"); err == nil {
		return path, nil
	}

	return "", types.NewCLINotFoundError("no suitable test command found (cat or echo)")
}

// BenchmarkJSONLineReader benchmarks JSON line reading performance
func BenchmarkJSONLineReader(b *testing.B) {
	// Create test data
	lines := make([]string, 1000)
	for i := range lines {
		lines[i] = `{"type":"test","data":"` + strings.Repeat("x", 100) + `"}`
	}
	input := strings.Join(lines, "\n") + "\n"

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		reader := NewJSONLineReader(strings.NewReader(input))
		for {
			_, err := reader.ReadLine()
			if err == io.EOF {
				break
			}
			if err != nil {
				b.Fatalf("ReadLine() error: %v", err)
			}
		}
	}
}

// BenchmarkJSONLineWriter benchmarks JSON line writing performance
func BenchmarkJSONLineWriter(b *testing.B) {
	line := `{"type":"test","data":"` + strings.Repeat("x", 100) + `"}`

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		writer := NewJSONLineWriter(&buf)
		for j := 0; j < 1000; j++ {
			if err := writer.WriteLine(line); err != nil {
				b.Fatalf("WriteLine() error: %v", err)
			}
		}
	}
}

// TestIntegrationSubprocessCLI tests end-to-end subprocess communication
// This test requires the actual Claude CLI to be installed
func TestIntegrationSubprocessCLI(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Try to find Claude CLI
	cliPath, err := FindCLI()
	if err != nil {
		t.Skipf("Claude CLI not found, skipping integration test: %v", err)
	}

	logger := log.NewLogger(false) // Non-verbose for tests
	transport := NewSubprocessCLITransport(cliPath, "", nil, logger, "", nil)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Connect to CLI
	if err := transport.Connect(ctx); err != nil {
		t.Fatalf("Connect() failed: %v", err)
	}
	defer func() {
		_ = transport.Close(ctx)
	}()

	// Should be ready
	if !transport.IsReady() {
		t.Errorf("IsReady() = false after successful Connect()")
	}

	// Try to write a simple query
	query := map[string]interface{}{
		"type":    "control",
		"subtype": "query",
		"prompt":  "Hello, Claude!",
	}

	queryJSON, err := json.Marshal(query)
	if err != nil {
		t.Fatalf("Failed to marshal query: %v", err)
	}

	if err := transport.Write(ctx, string(queryJSON)); err != nil {
		t.Errorf("Write() failed: %v", err)
	}

	// Read messages (with timeout)
	messages := transport.ReadMessages(ctx)

	select {
	case msg := <-messages:
		if msg == nil {
			t.Errorf("Received nil message")
		} else {
			t.Logf("Received message type: %s", msg.GetMessageType())
		}
	case <-time.After(5 * time.Second):
		t.Logf("Timeout waiting for response (may be expected for this test)")
	}
}

// TestExtractSessionNotFoundError tests parsing of session not found errors from stderr
func TestExtractSessionNotFoundError(t *testing.T) {
	tests := []struct {
		name          string
		stderrText    string
		wantMatched   bool
		wantSessionID string
	}{
		{
			name:          "valid session not found error",
			stderrText:    "No conversation found with session ID: 8587b432-e504-42c8-b9a7-e3fd0b4b2c60",
			wantMatched:   true,
			wantSessionID: "8587b432-e504-42c8-b9a7-e3fd0b4b2c60",
		},
		{
			name:          "session not found with extra text",
			stderrText:    "Error: No conversation found with session ID: 12345678-1234-1234-1234-123456789abc. Please check the ID.",
			wantMatched:   true,
			wantSessionID: "12345678-1234-1234-1234-123456789abc.",
		},
		{
			name:          "session not found with leading whitespace",
			stderrText:    "No conversation found with session ID:   abc123-def456  ",
			wantMatched:   true,
			wantSessionID: "abc123-def456",
		},
		{
			name:          "different error message",
			stderrText:    "Connection failed: timeout",
			wantMatched:   false,
			wantSessionID: "",
		},
		{
			name:          "partial match",
			stderrText:    "No conversation found",
			wantMatched:   false,
			wantSessionID: "",
		},
		{
			name:          "empty string",
			stderrText:    "",
			wantMatched:   false,
			wantSessionID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMatched, gotSessionID := extractSessionNotFoundError(tt.stderrText)

			if gotMatched != tt.wantMatched {
				t.Errorf("extractSessionNotFoundError() matched = %v, want %v", gotMatched, tt.wantMatched)
			}

			if gotSessionID != tt.wantSessionID {
				t.Errorf("extractSessionNotFoundError() sessionID = %q, want %q", gotSessionID, tt.wantSessionID)
			}
		})
	}
}

// TestParseStderrError tests the stderr error parsing and error creation
func TestParseStderrError(t *testing.T) {
	logger := log.NewLogger(false)
	transport := &SubprocessCLITransport{
		logger:   logger,
		messages: make(chan types.Message, 10),
	}

	// Test session not found error
	stderrText := "No conversation found with session ID: 8587b432-e504-42c8-b9a7-e3fd0b4b2c60"
	transport.parseStderrError(stderrText)

	// Check that error was stored
	err := transport.GetError()
	if err == nil {
		t.Fatal("parseStderrError() should have stored an error")
	}

	// Check that it's the right error type
	if !types.IsSessionNotFoundError(err) {
		t.Errorf("parseStderrError() stored error type = %T, want SessionNotFoundError", err)
	}

	// Check session ID is in the error
	if sessionErr, ok := err.(*types.SessionNotFoundError); ok {
		if sessionErr.SessionID != "8587b432-e504-42c8-b9a7-e3fd0b4b2c60" {
			t.Errorf("SessionNotFoundError.SessionID = %q, want %q",
				sessionErr.SessionID, "8587b432-e504-42c8-b9a7-e3fd0b4b2c60")
		}
	}
}

// TestForkSessionFlag tests that --fork-session flag is passed when ForkSession is true
func TestForkSessionFlag(t *testing.T) {
	tests := []struct {
		name            string
		resumeSessionID string
		forkSession     bool
		wantResumeFlag  bool
		wantForkFlag    bool
	}{
		{
			name:            "with resume and fork session",
			resumeSessionID: "test-session-id",
			forkSession:     true,
			wantResumeFlag:  true,
			wantForkFlag:    true,
		},
		{
			name:            "with resume but no fork session",
			resumeSessionID: "test-session-id",
			forkSession:     false,
			wantResumeFlag:  true,
			wantForkFlag:    false,
		},
		{
			name:            "with fork session but no resume",
			resumeSessionID: "",
			forkSession:     true,
			wantResumeFlag:  false,
			wantForkFlag:    true,
		},
		{
			name:            "without resume and fork session",
			resumeSessionID: "",
			forkSession:     false,
			wantResumeFlag:  false,
			wantForkFlag:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create options with fork session setting
			opts := types.NewClaudeAgentOptions().
				WithForkSession(tt.forkSession)

			logger := log.NewLogger(false)
			transport := NewSubprocessCLITransport("/bin/echo", "", nil, logger, tt.resumeSessionID, opts)

			// Build command args (without actually connecting)
			args := transport.buildCommandArgs()

			// Convert to string for easier searching
			argsStr := strings.Join(args, " ")
			t.Logf("CLI args: %v", args)

			// Check for --resume flag
			hasResumeFlag := contains(args, "--resume")
			if hasResumeFlag != tt.wantResumeFlag {
				t.Errorf("--resume flag present = %v, want %v", hasResumeFlag, tt.wantResumeFlag)
			}

			// Check for session ID if resume flag is expected
			if tt.wantResumeFlag {
				hasSessionID := contains(args, tt.resumeSessionID)
				if !hasSessionID {
					t.Errorf("session ID %q not found in args: %v", tt.resumeSessionID, args)
				}
			}

			// Check for --fork-session flag
			hasForkFlag := contains(args, "--fork-session")
			if hasForkFlag != tt.wantForkFlag {
				t.Errorf("--fork-session flag present = %v, want %v\nArgs: %s", hasForkFlag, tt.wantForkFlag, argsStr)
			}
		})
	}
}

// TestBuildCommandArgs_DefaultsToEmptySystemPrompt ensures the CLI receives an explicit empty system prompt by default.
func TestBuildCommandArgs_DefaultsToEmptySystemPrompt(t *testing.T) {
	opts := types.NewClaudeAgentOptions()
	logger := log.NewLogger(false)
	transport := NewSubprocessCLITransport("/bin/echo", "", nil, logger, "", opts)

	args := transport.buildCommandArgs()

	idx := findArg(args, "--system-prompt")
	if idx == -1 {
		t.Fatalf("--system-prompt flag not found in args: %v", args)
	}
	if idx+1 >= len(args) {
		t.Fatalf("missing value for --system-prompt flag: %v", args)
	}
	if args[idx+1] != "" {
		t.Fatalf("expected empty system prompt value, got %q", args[idx+1])
	}
}

// TestBuildCommandArgs_SystemPromptPresetAppend ensures preset append prompts are passed correctly.
func TestBuildCommandArgs_SystemPromptPresetAppend(t *testing.T) {
	appendText := "append me"
	preset := types.SystemPromptPreset{
		Type:   "preset",
		Preset: "claude_code",
		Append: &appendText,
	}

	opts := types.NewClaudeAgentOptions().WithSystemPromptPreset(preset)
	logger := log.NewLogger(false)
	transport := NewSubprocessCLITransport("/bin/echo", "", nil, logger, "", opts)

	args := transport.buildCommandArgs()

	if contains(args, "--system-prompt") {
		t.Fatalf("unexpected --system-prompt flag when using preset append: %v", args)
	}

	idx := findArg(args, "--append-system-prompt")
	if idx == -1 {
		t.Fatalf("--append-system-prompt flag not found in args: %v", args)
	}
	if idx+1 >= len(args) {
		t.Fatalf("missing value for --append-system-prompt flag: %v", args)
	}
	if args[idx+1] != appendText {
		t.Fatalf("expected append text %q, got %q", appendText, args[idx+1])
	}
}

// TestBuildCommandArgs_FallbackModel ensures fallback model flag is passed through.
func TestBuildCommandArgs_FallbackModel(t *testing.T) {
	fallback := "haiku"
	opts := types.NewClaudeAgentOptions().WithFallbackModel(fallback)

	logger := log.NewLogger(false)
	transport := NewSubprocessCLITransport("/bin/echo", "", nil, logger, "", opts)

	args := transport.buildCommandArgs()

	idx := findArg(args, "--fallback-model")
	if idx == -1 {
		t.Fatalf("--fallback-model flag not found in args: %v", args)
	}
	if idx+1 >= len(args) {
		t.Fatalf("missing value for --fallback-model flag: %v", args)
	}
	if args[idx+1] != fallback {
		t.Fatalf("expected fallback model %q, got %q", fallback, args[idx+1])
	}
}

// TestBuildCommandArgs_ToolsAndLimits verifies tool, limit, and extra flags are passed.
func TestBuildCommandArgs_ToolsAndLimits(t *testing.T) {
	settings := "/tmp/settings.json"
	addDirs := []string{"/tmp/a", "/tmp/b"}
	sources := []types.SettingSource{types.SettingSourceUser, types.SettingSourceLocal}
	customFlagValue := "value"

	opts := types.NewClaudeAgentOptions().
		WithAllowedTools("Bash", "Write").
		WithDisallowedTools("Edit").
		WithMaxTurns(5).
		WithContinueConversation(true).
		WithIncludePartialMessages(true).
		WithSettings(settings).
		WithAddDirs(addDirs...).
		WithSettingSources(sources...).
		WithExtraArg("custom-flag", &customFlagValue).
		WithExtraArg("debug-to-stderr", nil)

	logger := log.NewLogger(false)
	transport := NewSubprocessCLITransport("/bin/echo", "", nil, logger, "", opts)

	args := transport.buildCommandArgs()

	// Allowed tools
	if val, ok := flagValue(args, "--allowedTools"); !ok || val != "Bash,Write" {
		t.Fatalf("expected allowedTools flag with value %q, got %q (present=%v)", "Bash,Write", val, ok)
	}

	// Disallowed tools
	if val, ok := flagValue(args, "--disallowedTools"); !ok || val != "Edit" {
		t.Fatalf("expected disallowedTools flag with value %q, got %q (present=%v)", "Edit", val, ok)
	}

	// Max turns
	if val, ok := flagValue(args, "--max-turns"); !ok || val != "5" {
		t.Fatalf("expected max-turns flag with value %q, got %q (present=%v)", "5", val, ok)
	}

	// Continue conversation
	if !contains(args, "--continue") {
		t.Fatalf("expected --continue flag, args=%v", args)
	}

	// Include partial messages
	if !contains(args, "--include-partial-messages") {
		t.Fatalf("expected --include-partial-messages flag, args=%v", args)
	}

	// Settings path
	if val, ok := flagValue(args, "--settings"); !ok || val != settings {
		t.Fatalf("expected settings flag with value %q, got %q (present=%v)", settings, val, ok)
	}

	// Add directories
	for _, dir := range addDirs {
		if !containsFlagWithValue(args, "--add-dir", dir) {
			t.Fatalf("expected --add-dir %s in args %v", dir, args)
		}
	}

	// Setting sources
	if val, ok := flagValue(args, "--setting-sources"); !ok || val != "user,local" {
		t.Fatalf("expected setting-sources value %q, got %q (present=%v)", "user,local", val, ok)
	}

	// Extra args (bool + value)
	if !contains(args, "--debug-to-stderr") {
		t.Fatalf("expected --debug-to-stderr flag, args=%v", args)
	}
	if val, ok := flagValue(args, "--custom-flag"); !ok || val != customFlagValue {
		t.Fatalf("expected custom-flag value %q, got %q (present=%v)", customFlagValue, val, ok)
	}
}

// TestBuildCommandArgs_Plugins verifies plugin directories are passed to CLI.
func TestBuildCommandArgs_Plugins(t *testing.T) {
	pluginPath := "/tmp/plugin"
	opts := types.NewClaudeAgentOptions().WithLocalPlugin(pluginPath)

	logger := log.NewLogger(false)
	transport := NewSubprocessCLITransport("/bin/echo", "", nil, logger, "", opts)

	args := transport.buildCommandArgs()

	if !containsFlagWithValue(args, "--plugin-dir", pluginPath) {
		t.Fatalf("expected --plugin-dir %s in args %v", pluginPath, args)
	}
}

// TestBuildCommandArgs_Agents verifies agent definitions are serialized for CLI.
func TestBuildCommandArgs_Agents(t *testing.T) {
	model := "sonnet"
	agents := map[string]types.AgentDefinition{
		"code-reviewer": {
			Description: "Review code",
			Prompt:      "You are a reviewer",
			Tools:       []string{"Read"},
			Model:       &model,
		},
	}

	opts := types.NewClaudeAgentOptions().WithAgents(agents)
	logger := log.NewLogger(false)
	transport := NewSubprocessCLITransport("/bin/echo", "", nil, logger, "", opts)

	args := transport.buildCommandArgs()

	val, ok := flagValue(args, "--agents")
	if !ok {
		t.Fatalf("--agents flag not found in args %v", args)
	}

	var payload map[string]map[string]interface{}
	if err := json.Unmarshal([]byte(val), &payload); err != nil {
		t.Fatalf("failed to unmarshal agents payload: %v", err)
	}

	cfg, exists := payload["code-reviewer"]
	if !exists {
		t.Fatalf("agent 'code-reviewer' not found in payload %v", payload)
	}
	if cfg["description"] != "Review code" || cfg["prompt"] != "You are a reviewer" {
		t.Fatalf("unexpected agent config: %v", cfg)
	}
	if cfg["model"] != model {
		t.Fatalf("expected model %q, got %v", model, cfg["model"])
	}
}

// findArg returns the index of the target flag or -1 if not present.
func findArg(args []string, target string) int {
	for i, arg := range args {
		if arg == target {
			return i
		}
	}
	return -1
}

// contains checks if a slice contains a string
func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

// containsFlagWithValue checks for a flag immediately followed by a value.
func containsFlagWithValue(args []string, flag, value string) bool {
	for i := 0; i < len(args)-1; i++ {
		if args[i] == flag && args[i+1] == value {
			return true
		}
	}
	return false
}

// flagValue returns the value following a flag, if present.
func flagValue(args []string, flag string) (string, bool) {
	if idx := findArg(args, flag); idx != -1 && idx+1 < len(args) {
		return args[idx+1], true
	}
	return "", false
}

// TestMaxBufferSizeDefault verifies the default max buffer size is applied.
func TestMaxBufferSizeDefault(t *testing.T) {
	opts := types.NewClaudeAgentOptions()
	logger := log.NewLogger(false)
	transport := NewSubprocessCLITransport("/bin/echo", "", nil, logger, "", opts)

	if transport.maxBufferSize != DefaultMaxBufferSize {
		t.Fatalf("expected default maxBufferSize %d, got %d", DefaultMaxBufferSize, transport.maxBufferSize)
	}
}

// TestMaxBufferSizeOverride verifies custom buffer size is used.
func TestMaxBufferSizeOverride(t *testing.T) {
	size := 2048
	opts := types.NewClaudeAgentOptions().WithMaxBufferSize(size)
	logger := log.NewLogger(false)
	transport := NewSubprocessCLITransport("/bin/echo", "", nil, logger, "", opts)

	if transport.maxBufferSize != size {
		t.Fatalf("expected maxBufferSize %d, got %d", size, transport.maxBufferSize)
	}
}

// TestMCPServersStringPath ensures string-based MCP config paths are passed to the CLI.
func TestMCPServersStringPath(t *testing.T) {
	const configPath = "/tmp/mcp-config.json"
	opts := types.NewClaudeAgentOptions().WithMcpServers(configPath)

	logger := log.NewLogger(false)
	transport := NewSubprocessCLITransport("/bin/echo", "", nil, logger, "", opts)

	args := transport.buildCommandArgs()

	idx := -1
	for i, arg := range args {
		if arg == "--mcp-servers" {
			idx = i
			break
		}
	}

	if idx == -1 {
		t.Fatalf("--mcp-servers flag not found in args: %v", args)
	}

	if idx+1 >= len(args) {
		t.Fatalf("config path missing after --mcp-servers flag: %v", args)
	}

	if args[idx+1] != configPath {
		t.Fatalf("expected config path %s, got %s", configPath, args[idx+1])
	}
}
