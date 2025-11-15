package transport

import (
	"testing"
)

func TestParseSemanticVersion(t *testing.T) {
	tests := []struct {
		name        string
		versionStr  string
		expected    SemanticVersion
		expectError bool
	}{
		{
			name:       "valid version 2.0.0",
			versionStr: "2.0.0",
			expected:   SemanticVersion{Major: 2, Minor: 0, Patch: 0},
		},
		{
			name:       "valid version 2.1.5",
			versionStr: "2.1.5",
			expected:   SemanticVersion{Major: 2, Minor: 1, Patch: 5},
		},
		{
			name:       "version with leading v",
			versionStr: "v2.3.4",
			expected:   SemanticVersion{Major: 2, Minor: 3, Patch: 4},
		},
		{
			name:       "version with whitespace",
			versionStr: "  2.1.0  ",
			expected:   SemanticVersion{Major: 2, Minor: 1, Patch: 0},
		},
		{
			name:       "version with pre-release",
			versionStr: "2.1.0-beta.1",
			expected:   SemanticVersion{Major: 2, Minor: 1, Patch: 0},
		},
		{
			name:        "invalid version - missing parts",
			versionStr:  "2.1",
			expectError: true,
		},
		{
			name:        "invalid version - non-numeric",
			versionStr:  "a.b.c",
			expectError: true,
		},
		{
			name:        "invalid version - empty",
			versionStr:  "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version, err := ParseSemanticVersion(tt.versionStr)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error parsing '%s', got none", tt.versionStr)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error parsing '%s': %v", tt.versionStr, err)
				return
			}

			if version != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, version)
			}
		})
	}
}

func TestSemanticVersionIsAtLeast(t *testing.T) {
	tests := []struct {
		name     string
		version  SemanticVersion
		required SemanticVersion
		expected bool
	}{
		{
			name:     "equal versions",
			version:  SemanticVersion{Major: 2, Minor: 0, Patch: 0},
			required: SemanticVersion{Major: 2, Minor: 0, Patch: 0},
			expected: true,
		},
		{
			name:     "newer major version",
			version:  SemanticVersion{Major: 3, Minor: 0, Patch: 0},
			required: SemanticVersion{Major: 2, Minor: 0, Patch: 0},
			expected: true,
		},
		{
			name:     "newer minor version",
			version:  SemanticVersion{Major: 2, Minor: 1, Patch: 0},
			required: SemanticVersion{Major: 2, Minor: 0, Patch: 0},
			expected: true,
		},
		{
			name:     "newer patch version",
			version:  SemanticVersion{Major: 2, Minor: 0, Patch: 5},
			required: SemanticVersion{Major: 2, Minor: 0, Patch: 0},
			expected: true,
		},
		{
			name:     "older major version",
			version:  SemanticVersion{Major: 1, Minor: 9, Patch: 9},
			required: SemanticVersion{Major: 2, Minor: 0, Patch: 0},
			expected: false,
		},
		{
			name:     "older minor version",
			version:  SemanticVersion{Major: 2, Minor: 0, Patch: 0},
			required: SemanticVersion{Major: 2, Minor: 1, Patch: 0},
			expected: false,
		},
		{
			name:     "older patch version",
			version:  SemanticVersion{Major: 2, Minor: 0, Patch: 0},
			required: SemanticVersion{Major: 2, Minor: 0, Patch: 1},
			expected: false,
		},
		{
			name:     "complex: newer minor, same major",
			version:  SemanticVersion{Major: 2, Minor: 5, Patch: 0},
			required: SemanticVersion{Major: 2, Minor: 1, Patch: 10},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.version.IsAtLeast(tt.required)
			if result != tt.expected {
				t.Errorf("Expected %v.IsAtLeast(%v) to be %v, got %v",
					tt.version, tt.required, tt.expected, result)
			}
		})
	}
}

func TestSemanticVersionString(t *testing.T) {
	tests := []struct {
		version  SemanticVersion
		expected string
	}{
		{
			version:  SemanticVersion{Major: 2, Minor: 0, Patch: 0},
			expected: "2.0.0",
		},
		{
			version:  SemanticVersion{Major: 1, Minor: 2, Patch: 3},
			expected: "1.2.3",
		},
		{
			version:  SemanticVersion{Major: 10, Minor: 20, Patch: 30},
			expected: "10.20.30",
		},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.version.String()
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}
