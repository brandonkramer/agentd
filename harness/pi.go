package harness

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/brandonkramer/shellquote"
)

const (
	piHarness      = "pi"
	piExecutable   = "pi"
	piNoSession    = "--no-session"
	piNoExtensions = "--no-extensions"
	piDisableTools = "-nt"
)

type piDriver struct{}

func (piDriver) Name() string { return Pi }

// PiBaseArgs returns default non-interactive Pi CLI flags for agentd dispatch.
func PiBaseArgs() []string {
	return []string{"--print", "--mode", "json", piNoSession, piNoExtensions, piDisableTools}
}

func (piDriver) Prepare(in *WorkInput) (Prepared, error) {
	piPath, err := FindPiCLI()
	if err != nil {
		return Prepared{}, err
	}
	promptPath, err := shellquote.WritePrompt(in.RunDir, in.PromptContent)
	if err != nil {
		return Prepared{}, fmt.Errorf("pi harness: write prompt: %w", err)
	}
	args, warnings := piArgs(in.Model, promptPath)
	display := formatExecDisplay(piPath, args)
	return Prepared{
		Driver: Pi, Harness: piHarness, Protocol: ProtocolPiJSON,
		CommandTemplate: piTemplate(in.Model),
		Command:         display, PromptPath: promptPath,
		ExecPath: piPath, ExecArgs: args, ExecDir: in.WorkDir,
		Warnings: warnings,
	}, nil
}

// FindPiCLI locates the pi executable on PATH.
func FindPiCLI() (string, error) {
	path, err := exec.LookPath(piExecutable)
	if err != nil {
		return "", fmt.Errorf("pi driver: %s not found in PATH (install Pi: https://pi.dev)", piExecutable)
	}
	return path, nil
}

func piArgs(model, promptPath string) (args, warnings []string) {
	args = append([]string(nil), PiBaseArgs()...)
	if model != "" {
		args = append(args, "--model", model)
	}
	args = append(args, "@"+promptPath)
	return args, warnings
}

func piTemplate(model string) string {
	parts := append([]string{piExecutable}, PiBaseArgs()...)
	if model != "" {
		parts = append(parts, "--model", model)
	}
	parts = append(parts, "@"+shellquote.DefaultPromptFile)
	return strings.Join(parts, " ")
}

// PreparePi prepares a pi headless work unit.
func PreparePi(in *WorkInput) (Prepared, error) {
	return piDriver{}.Prepare(in)
}
