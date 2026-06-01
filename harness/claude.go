package harness

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/brandonkramer/shellquote"
)

const (
	claudeHarness              = "claude-code"
	claudeNoSessionPersistence = "--no-session-persistence"
)

// ClaudeBaseArgs returns default non-interactive Claude CLI flags.
func ClaudeBaseArgs() []string {
	return []string{"-p", "--output-format", "text", claudeNoSessionPersistence}
}

type claudeCode struct{}

func (claudeCode) Name() string { return ClaudeCode }

func (claudeCode) Prepare(in *WorkInput) (Prepared, error) {
	claudePath, err := FindClaudeCLI()
	if err != nil {
		return Prepared{}, err
	}
	promptPath, err := shellquote.WritePrompt(in.RunDir, in.PromptContent)
	if err != nil {
		return Prepared{}, fmt.Errorf("claude-code harness: write prompt: %w", err)
	}
	args, warnings := claudeArgs(in.Model)
	display := formatExecDisplay(claudePath, args) + " < " + promptPath
	return Prepared{
		Driver: ClaudeCode, Harness: claudeHarness,
		CommandTemplate: claudeTemplate(in.Model),
		Command:         display, PromptPath: promptPath,
		ExecPath: claudePath, ExecArgs: args, ExecDir: in.WorkDir,
		StdinPrompt: true, Warnings: warnings,
	}, nil
}

// FindClaudeCLI locates the claude executable on PATH.
func FindClaudeCLI() (string, error) {
	path, err := exec.LookPath("claude")
	if err != nil {
		return "", fmt.Errorf("claude-code driver: claude CLI not found in PATH")
	}
	return path, nil
}

func claudeArgs(model string) (args, warnings []string) {
	args = ClaudeBaseArgs()
	if model != "" {
		args = append(args, "--model", model)
	}
	return args, warnings
}

func claudeTemplate(model string) string {
	parts := []string{"claude", "-p", "--output-format", "text", claudeNoSessionPersistence}
	if model != "" {
		parts = append(parts, "--model", model)
	}
	parts = append(parts, "<", shellquote.DefaultPromptFile)
	return strings.Join(parts, " ")
}

func formatExecDisplay(path string, args []string) string {
	parts := append([]string{path}, args...)
	return strings.Join(parts, " ")
}

// PrepareClaude prepares a claude-code work unit.
func PrepareClaude(in *WorkInput) (Prepared, error) {
	return claudeCode{}.Prepare(in)
}
