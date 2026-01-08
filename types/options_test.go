package types

import (
	"testing"
)

// TestWithMaxThinkingTokens tests the WithMaxThinkingTokens builder method.
func TestWithMaxThinkingTokens(t *testing.T) {
	opts := NewClaudeAgentOptions()

	// Test setting max thinking tokens
	result := opts.WithMaxThinkingTokens(5000)

	// Verify the method returns the same instance for chaining
	if result != opts {
		t.Error("WithMaxThinkingTokens should return the same instance for chaining")
	}

	// Verify the value is set correctly
	if opts.MaxThinkingTokens == nil {
		t.Fatal("MaxThinkingTokens should not be nil after setting")
	}

	if *opts.MaxThinkingTokens != 5000 {
		t.Errorf("Expected MaxThinkingTokens to be 5000, got %d", *opts.MaxThinkingTokens)
	}
}

// TestWithMaxBudgetUSD tests the WithMaxBudgetUSD builder method.
func TestWithMaxBudgetUSD(t *testing.T) {
	opts := NewClaudeAgentOptions()

	// Test setting max budget
	result := opts.WithMaxBudgetUSD(10.50)

	// Verify the method returns the same instance for chaining
	if result != opts {
		t.Error("WithMaxBudgetUSD should return the same instance for chaining")
	}

	// Verify the value is set correctly
	if opts.MaxBudgetUSD == nil {
		t.Fatal("MaxBudgetUSD should not be nil after setting")
	}

	if *opts.MaxBudgetUSD != 10.50 {
		t.Errorf("Expected MaxBudgetUSD to be 10.50, got %.2f", *opts.MaxBudgetUSD)
	}
}

// TestOptionsChaining tests that the builder methods can be chained together.
func TestOptionsChaining(t *testing.T) {
	opts := NewClaudeAgentOptions().
		WithMaxThinkingTokens(8000).
		WithMaxBudgetUSD(25.00).
		WithModel("claude-3-5-sonnet-20241022").
		WithMaxTurns(10)

	// Verify all values are set correctly
	if opts.MaxThinkingTokens == nil || *opts.MaxThinkingTokens != 8000 {
		t.Error("MaxThinkingTokens not set correctly in chain")
	}

	if opts.MaxBudgetUSD == nil || *opts.MaxBudgetUSD != 25.00 {
		t.Error("MaxBudgetUSD not set correctly in chain")
	}

	if opts.Model == nil || *opts.Model != "claude-3-5-sonnet-20241022" {
		t.Error("Model not set correctly in chain")
	}

	if opts.MaxTurns == nil || *opts.MaxTurns != 10 {
		t.Error("MaxTurns not set correctly in chain")
	}
}

// TestNewClaudeAgentOptions tests that the constructor creates a valid instance.
func TestNewClaudeAgentOptions(t *testing.T) {
	opts := NewClaudeAgentOptions()

	// Verify default values
	if opts == nil {
		t.Fatal("NewClaudeAgentOptions should not return nil")
	}

	// Check that optional fields are nil by default
	if opts.MaxThinkingTokens != nil {
		t.Error("MaxThinkingTokens should be nil by default")
	}

	if opts.MaxBudgetUSD != nil {
		t.Error("MaxBudgetUSD should be nil by default")
	}

	// Check that maps are initialized
	if opts.Env == nil {
		t.Error("Env should be initialized")
	}

	if opts.ExtraArgs == nil {
		t.Error("ExtraArgs should be initialized")
	}
}

// TestWithMaxThinkingTokensZeroValue tests that zero values can be set.
func TestWithMaxThinkingTokensZeroValue(t *testing.T) {
	opts := NewClaudeAgentOptions().WithMaxThinkingTokens(0)

	if opts.MaxThinkingTokens == nil {
		t.Fatal("MaxThinkingTokens should not be nil")
	}

	if *opts.MaxThinkingTokens != 0 {
		t.Errorf("Expected MaxThinkingTokens to be 0, got %d", *opts.MaxThinkingTokens)
	}
}

// TestWithMaxBudgetUSDZeroValue tests that zero budget can be set.
func TestWithMaxBudgetUSDZeroValue(t *testing.T) {
	opts := NewClaudeAgentOptions().WithMaxBudgetUSD(0.0)

	if opts.MaxBudgetUSD == nil {
		t.Fatal("MaxBudgetUSD should not be nil")
	}

	if *opts.MaxBudgetUSD != 0.0 {
		t.Errorf("Expected MaxBudgetUSD to be 0.0, got %.2f", *opts.MaxBudgetUSD)
	}
}

// TestWithTools validates base tools configuration.
func TestWithTools(t *testing.T) {
	opts := NewClaudeAgentOptions().WithTools("Read", "Write")

	tools, ok := opts.Tools.([]string)
	if !ok {
		t.Fatalf("expected Tools to be []string, got %T", opts.Tools)
	}

	if len(tools) != 2 || tools[0] != "Read" || tools[1] != "Write" {
		t.Fatalf("unexpected tools slice: %v", tools)
	}
}

// TestWithBetas validates beta headers configuration.
func TestWithBetas(t *testing.T) {
	opts := NewClaudeAgentOptions().WithBetas(SdkBetaContext1M)

	if len(opts.Betas) != 1 || opts.Betas[0] != SdkBetaContext1M {
		t.Fatalf("expected betas to contain %s, got %v", SdkBetaContext1M, opts.Betas)
	}
}

// TestWithJSONSchemaOutput validates structured output helper.
func TestWithJSONSchemaOutput(t *testing.T) {
	schema := map[string]interface{}{
		"type": "object",
	}

	opts := NewClaudeAgentOptions().WithJSONSchemaOutput(schema)
	if opts.OutputFormat == nil {
		t.Fatalf("expected OutputFormat to be set")
	}

	if opts.OutputFormat["type"] != "json_schema" {
		t.Fatalf("expected type to be json_schema, got %v", opts.OutputFormat["type"])
	}
}

// TestWithEnableFileCheckpointing toggles the flag.
func TestWithEnableFileCheckpointing(t *testing.T) {
	opts := NewClaudeAgentOptions().WithEnableFileCheckpointing(true)
	if !opts.EnableFileCheckpointing {
		t.Fatalf("expected EnableFileCheckpointing to be true")
	}
}
