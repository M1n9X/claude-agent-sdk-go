package transport

import (
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/M1n9X/claude-agent-sdk-go/types"
)

// FindCLI searches for Claude Code CLI binary in standard locations.
// It checks in this order:
//  1. PATH via exec.LookPath("claude")
//  2. Default Claude installation location (new in CLI 2.0+):
//     - ~/.claude/local/claude
//  3. Common npm/yarn global install locations:
//     - ~/.npm-global/bin/claude
//     - /usr/local/bin/claude
//     - ~/.local/bin/claude
//     - ~/node_modules/.bin/claude
//     - ~/.yarn/bin/claude
//
// After finding the CLI, it checks the version to ensure it meets minimum requirements
// (unless CLAUDE_AGENT_SDK_SKIP_VERSION_CHECK is set).
//
// Returns the path to the CLI binary or a CLINotFoundError if not found.
func FindCLI() (string, error) {
	// First, try to find in PATH
	if cliPath, err := exec.LookPath("claude"); err == nil {
		// Check version before returning
		if err := CheckCLIVersion(cliPath); err != nil {
			return "", err
		}
		return cliPath, nil
	}

	// Try common install locations
	locations := []string{
		"~/.claude/local/claude", // Default location (CLI 2.0+)
		"~/.npm-global/bin/claude",
		"/usr/local/bin/claude",
		"~/.local/bin/claude",
		"~/node_modules/.bin/claude",
		"~/.yarn/bin/claude",
	}

	for _, location := range locations {
		expandedPath := expandHome(location)
		if _, err := os.Stat(expandedPath); err == nil {
			// Check version before returning
			if err := CheckCLIVersion(expandedPath); err != nil {
				return "", err
			}
			return expandedPath, nil
		}
	}

	// Not found anywhere
	return "", types.NewCLINotFoundError(
		"Claude Code not found. Install with:\n" +
			"  npm install -g @anthropic-ai/claude-code\n" +
			"\nIf already installed locally, try:\n" +
			"  export PATH=\"$HOME/node_modules/.bin:$PATH\"\n" +
			"\nOr provide the path via ClaudeAgentOptions:\n" +
			"  ClaudeAgentOptions{CLIPath: \"/path/to/claude\"}",
	)
}

// expandHome expands the ~ prefix in a path to the user's home directory.
// If the path does not start with ~, it is returned unchanged.
// If the home directory cannot be determined, the path is returned unchanged.
func expandHome(path string) string {
	if !strings.HasPrefix(path, "~") {
		return path
	}

	// Get current user
	usr, err := user.Current()
	if err != nil {
		return path
	}

	// Replace ~ with home directory
	if path == "~" {
		return usr.HomeDir
	}
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(usr.HomeDir, path[2:])
	}

	return path
}
