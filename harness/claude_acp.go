package harness

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/brandonkramer/shellquote"
)

const (
	claudeACPHarness    = "claude-code"
	claudeACPExecutable = "claude-code-acp"
)

type claudeACP struct{}

func (claudeACP) Name() string { return ClaudeACP }

func (claudeACP) Prepare(in *WorkInput) (Prepared, error) {
	acpPath, err := FindClaudeACPCLI()
	if err != nil {
		return Prepared{}, err
	}
	promptPath, err := shellquote.WritePrompt(in.RunDir, in.PromptContent)
	if err != nil {
		return Prepared{}, fmt.Errorf("claude-acp harness: write prompt: %w", err)
	}
	args, warnings := claudeACPArgs(in.Model)
	display := formatExecDisplay(acpPath, args)
	return Prepared{
		Driver: ClaudeACP, Harness: claudeACPHarness, Protocol: ProtocolACP,
		CommandTemplate: claudeACPTemplate(in.Model),
		Command:         display, PromptPath: promptPath,
		ExecPath: acpPath, ExecArgs: args, ExecDir: in.WorkDir,
		Warnings: warnings,
	}, nil
}

// FindClaudeACPCLI locates the claude-code-acp executable on PATH.
func FindClaudeACPCLI() (string, error) {
	path, err := exec.LookPath(claudeACPExecutable)
	if err != nil {
		return "", fmt.Errorf("claude-acp driver: %s not found in PATH (install with `npm install -g @zed-industries/claude-code-acp`)", claudeACPExecutable)
	}
	return path, nil
}

func claudeACPArgs(model string) (args, warnings []string) {
	if model != "" {
		warnings = append(warnings, "claude-acp: model selection is not yet passed to the ACP server; configure the model in Claude Code settings")
	}
	return args, warnings
}

func claudeACPTemplate(model string) string {
	parts := []string{claudeACPExecutable}
	if model != "" {
		parts = append(parts, "# model="+model)
	}
	return strings.Join(parts, " ")
}

// PrepareClaudeACP prepares a claude-acp work unit.
func PrepareClaudeACP(in *WorkInput) (Prepared, error) {
	return claudeACP{}.Prepare(in)
}
