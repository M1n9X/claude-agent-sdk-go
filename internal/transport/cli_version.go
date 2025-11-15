package transport

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/M1n9X/claude-agent-sdk-go/types"
)

const (
	// MinimumCLIVersion is the minimum required CLI version (2.0.0)
	MinimumCLIVersion = "2.0.0"

	// MinimumCLIMajor is the minimum major version
	MinimumCLIMajor = 2

	// MinimumCLIMinor is the minimum minor version
	MinimumCLIMinor = 0

	// MinimumCLIPatch is the minimum patch version
	MinimumCLIPatch = 0
)

// SemanticVersion represents a semantic version number (major.minor.patch)
type SemanticVersion struct {
	Major int
	Minor int
	Patch int
}

// String returns the string representation of the version
func (v SemanticVersion) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// IsAtLeast checks if this version is at least the specified version
func (v SemanticVersion) IsAtLeast(required SemanticVersion) bool {
	if v.Major > required.Major {
		return true
	}
	if v.Major < required.Major {
		return false
	}

	// Major versions are equal, check minor
	if v.Minor > required.Minor {
		return true
	}
	if v.Minor < required.Minor {
		return false
	}

	// Major and minor are equal, check patch
	return v.Patch >= required.Patch
}

// ParseSemanticVersion parses a semantic version string (e.g., "2.1.0")
func ParseSemanticVersion(versionStr string) (SemanticVersion, error) {
	// Clean the version string (remove leading 'v' if present)
	versionStr = strings.TrimSpace(versionStr)
	versionStr = strings.TrimPrefix(versionStr, "v")

	// Use regex to extract version numbers
	// Pattern: major.minor.patch with optional pre-release/metadata
	re := regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)`)
	matches := re.FindStringSubmatch(versionStr)

	if len(matches) != 4 {
		return SemanticVersion{}, fmt.Errorf("invalid version format: %s", versionStr)
	}

	major, err := strconv.Atoi(matches[1])
	if err != nil {
		return SemanticVersion{}, fmt.Errorf("invalid major version: %s", matches[1])
	}

	minor, err := strconv.Atoi(matches[2])
	if err != nil {
		return SemanticVersion{}, fmt.Errorf("invalid minor version: %s", matches[2])
	}

	patch, err := strconv.Atoi(matches[3])
	if err != nil {
		return SemanticVersion{}, fmt.Errorf("invalid patch version: %s", matches[3])
	}

	return SemanticVersion{
		Major: major,
		Minor: minor,
		Patch: patch,
	}, nil
}

// GetCLIVersion retrieves the version of the Claude CLI binary
func GetCLIVersion(cliPath string) (SemanticVersion, error) {
	// Create context with timeout to prevent hanging
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Run: claude --version
	cmd := exec.CommandContext(ctx, cliPath, "--version")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return SemanticVersion{}, fmt.Errorf("failed to get CLI version: %w (stderr: %s)", err, stderr.String())
	}

	// Parse version from output
	versionStr := strings.TrimSpace(stdout.String())
	return ParseSemanticVersion(versionStr)
}

// CheckCLIVersion verifies that the CLI version meets minimum requirements
// Returns nil if version is acceptable, or an error if not
func CheckCLIVersion(cliPath string) error {
	// Check if version checking is disabled via environment variable
	if os.Getenv("CLAUDE_AGENT_SDK_SKIP_VERSION_CHECK") != "" {
		return nil
	}

	// Get the CLI version
	version, err := GetCLIVersion(cliPath)
	if err != nil {
		// If we can't determine the version, warn but don't fail
		// (for backwards compatibility with older CLIs that might not have --version)
		return nil
	}

	// Check minimum version requirement
	minVersion := SemanticVersion{
		Major: MinimumCLIMajor,
		Minor: MinimumCLIMinor,
		Patch: MinimumCLIPatch,
	}

	if !version.IsAtLeast(minVersion) {
		return types.NewCLINotFoundError(fmt.Sprintf(
			"Claude CLI version %s is installed, but version %s or higher is required.\n"+
				"Please update with:\n"+
				"  npm install -g @anthropic-ai/claude-code@latest\n"+
				"\nTo skip this check, set:\n"+
				"  export CLAUDE_AGENT_SDK_SKIP_VERSION_CHECK=1",
			version.String(),
			minVersion.String(),
		))
	}

	return nil
}
