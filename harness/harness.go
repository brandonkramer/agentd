package harness

import (
	"fmt"
	"os/exec"
	"sort"
	"strings"

	"github.com/brandonkramer/registry"
)

// BuildInput describes a managed agent or dry-run validation request.
type BuildInput struct {
	WorkDir string
	Prompt  string
	Model   string
}

// Harness maps a registry name to command construction and run preparation.
type Harness struct {
	Name        string
	Build       func(in BuildInput) (bin string, args []string, err error)
	Prepare     func(in *WorkInput) (Prepared, error)
	Executables []string
}

var reg = registry.New[Harness](
	registry.WithValidator(func(h Harness) error {
		if h.Name == "" || h.Prepare == nil {
			return fmt.Errorf("harness: invalid registration")
		}
		return nil
	}),
	registry.WithKeyFrom(func(h Harness) string { return h.Name }),
)

// Register adds a harness to the global registry.
func Register(h Harness) { reg.MustRegisterItem(h) }

// Names returns registered harness names in sorted order.
func Names() []string { return reg.Names() }

// Get returns the harness registered under name.
func Get(name string) (Harness, error) {
	h, err := reg.Get(name)
	if err != nil {
		return Harness{}, fmt.Errorf("unknown harness %q", name)
	}
	return h, nil
}

// ExecutableNames returns sorted unique executable basenames required by registered harnesses.
func ExecutableNames() []string {
	seen := map[string]struct{}{}
	for _, h := range reg.Values() {
		for _, exe := range h.Executables {
			seen[exe] = struct{}{}
		}
	}
	out := make([]string, 0, len(seen))
	for exe := range seen {
		out = append(out, exe)
	}
	sort.Strings(out)
	return out
}

func init() {
	Register(Harness{
		Name: GenericCommand,
		Build: func(in BuildInput) (string, []string, error) {
			if strings.TrimSpace(in.Prompt) == "" {
				return "", nil, errCommandRequired
			}
			return "sh", []string{"-c", in.Prompt}, nil
		},
		Prepare:     PrepareGeneric,
		Executables: []string{"sh", "bash", "zsh"},
	})
	Register(Harness{
		Name: ClaudeCode,
		Build: func(in BuildInput) (string, []string, error) {
			path, err := exec.LookPath("claude")
			if err != nil {
				return "", nil, fmt.Errorf("claude-code harness: claude CLI not found in PATH: %w", err)
			}
			args := ClaudeBaseArgs()
			if strings.TrimSpace(in.Model) != "" {
				args = append(args, "--model", in.Model)
			}
			return path, args, nil
		},
		Prepare:     PrepareClaude,
		Executables: []string{"claude"},
	})
	Register(Harness{
		Name: ClaudeACP,
		Build: func(in BuildInput) (string, []string, error) {
			path, err := FindClaudeACPCLI()
			if err != nil {
				return "", nil, err
			}
			args, _ := claudeACPArgs(in.Model)
			return path, args, nil
		},
		Prepare:     PrepareClaudeACP,
		Executables: []string{claudeACPExecutable},
	})
	Register(Harness{
		Name: Pi,
		Build: func(in BuildInput) (string, []string, error) {
			path, err := FindPiCLI()
			if err != nil {
				return "", nil, err
			}
			if strings.TrimSpace(in.Prompt) == "" {
				return "", nil, errCommandRequired
			}
			args, _ := piArgs(in.Model, in.Prompt)
			return path, args, nil
		},
		Prepare:     PreparePi,
		Executables: []string{piExecutable},
	})
}

var errCommandRequired = errRequired("command is required")

type errRequired string

func (e errRequired) Error() string { return string(e) }
