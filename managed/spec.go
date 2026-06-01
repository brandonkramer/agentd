package managed

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/brandonkramer/catalogfile"
)

// ValidateDriver checks driver names when loading specs. Nil skips validation.
var ValidateDriver func(name string) error

// SetDriverValidator configures harness name validation during spec load.
func SetDriverValidator(fn func(name string) error) {
	ValidateDriver = fn
}

const (
	// DefaultReconcilerInterval is the default managed-agent reconcile loop period.
	DefaultReconcilerInterval = 5 * time.Second
	defaultDriverName         = "generic-command"
)

// AgentSpec is the resolved desired state for one managed agent.
type AgentSpec struct {
	Name         string
	Enabled      bool
	TaskID       string
	Title        string
	Body         string
	Driver       string
	WorkDir      string
	Prompt       string
	PromptFile   string
	Restart      string
	MaxRestarts  int
	Backoff      time.Duration
	Timeout      time.Duration
	GracePeriod  time.Duration
	PollInterval time.Duration
	Env          map[string]string
	SourcePath   string
}

type agentsFile struct {
	Agents []agentEntry `toml:"agents"`
}

type agentEntry struct {
	Name         string            `toml:"name"`
	Enabled      *bool             `toml:"enabled"`
	TaskID       string            `toml:"task_id"`
	Title        string            `toml:"title"`
	Body         string            `toml:"body"`
	Driver       string            `toml:"driver"`
	WorkDir      string            `toml:"work_dir"`
	Prompt       string            `toml:"prompt"`
	PromptFile   string            `toml:"prompt_file"`
	Restart      string            `toml:"restart"`
	MaxRestarts  int               `toml:"max_restarts"`
	Backoff      string            `toml:"backoff"`
	Timeout      string            `toml:"timeout"`
	GracePeriod  string            `toml:"grace_period"`
	PollInterval string            `toml:"poll_interval"`
	Env          map[string]string `toml:"env"`
}

// LoadMergedSpecs loads global and optional project agents.toml with project override by name.
func LoadMergedSpecs(home, projectRoot string, store Store) (map[string]AgentSpec, error) {
	return catalogfile.MergeCatalog(home, projectRoot, store, loadFile)
}

func loadFile(path, home, projectRoot string) (map[string]AgentSpec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var file agentsFile
	if err := toml.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	out := map[string]AgentSpec{}
	for i := range file.Agents {
		spec, err := normalizeEntry(&file.Agents[i], path, projectRoot)
		if err != nil {
			return nil, err
		}
		if _, exists := out[spec.Name]; exists {
			return nil, fmt.Errorf("duplicate managed agent %q in %s", spec.Name, path)
		}
		out[spec.Name] = spec
	}
	return out, nil
}

func normalizeEntry(entry *agentEntry, sourcePath, projectRoot string) (AgentSpec, error) {
	name := strings.TrimSpace(entry.Name)
	if name == "" {
		return AgentSpec{}, fmt.Errorf("managed agent name is required in %s", sourcePath)
	}
	enabled := true
	if entry.Enabled != nil {
		enabled = *entry.Enabled
	}
	workDir := strings.TrimSpace(entry.WorkDir)
	if workDir == "" && projectRoot != "" {
		workDir = projectRoot
	}
	if workDir != "" {
		abs, err := filepath.Abs(workDir)
		if err != nil {
			return AgentSpec{}, fmt.Errorf("agent %q: resolve work_dir: %w", name, err)
		}
		workDir = abs
	}
	restart := strings.TrimSpace(entry.Restart)
	if restart == "" {
		restart = "no"
	}
	switch restart {
	case "no", "on-failure", "always":
	default:
		return AgentSpec{}, fmt.Errorf("agent %q: invalid restart %q", name, restart)
	}
	backoff, err := parseDuration(entry.Backoff, 3*time.Second)
	if err != nil {
		return AgentSpec{}, fmt.Errorf("agent %q: %w", name, err)
	}
	timeout, err := parseOptionalDuration(entry.Timeout)
	if err != nil {
		return AgentSpec{}, fmt.Errorf("agent %q: %w", name, err)
	}
	grace, err := parseDuration(entry.GracePeriod, 30*time.Second)
	if err != nil {
		return AgentSpec{}, fmt.Errorf("agent %q: %w", name, err)
	}
	poll, err := parseDuration(entry.PollInterval, 2*time.Second)
	if err != nil {
		return AgentSpec{}, fmt.Errorf("agent %q: %w", name, err)
	}
	driverName := strings.TrimSpace(entry.Driver)
	if driverName == "" {
		driverName = defaultDriverName
	}
	if ValidateDriver != nil {
		if err := ValidateDriver(driverName); err != nil {
			return AgentSpec{}, fmt.Errorf("agent %q: %w", name, err)
		}
	}
	if entry.TaskID != "" && (entry.Title != "" || entry.Body != "") {
		return AgentSpec{}, fmt.Errorf("agent %q: task_id is mutually exclusive with title/body", name)
	}
	if entry.TaskID == "" && entry.Title == "" && entry.Body == "" && entry.Prompt == "" && entry.PromptFile == "" {
		return AgentSpec{}, fmt.Errorf("agent %q: no work source (task_id, title/body, prompt, or prompt_file)", name)
	}
	promptFile := entry.PromptFile
	if promptFile != "" && !filepath.IsAbs(promptFile) {
		promptFile = filepath.Join(filepath.Dir(sourcePath), promptFile)
	}
	return AgentSpec{
		Name: name, Enabled: enabled, TaskID: strings.TrimSpace(entry.TaskID),
		Title: entry.Title, Body: entry.Body, Driver: driverName, WorkDir: workDir,
		Prompt: entry.Prompt, PromptFile: promptFile, Restart: restart,
		MaxRestarts: entry.MaxRestarts, Backoff: backoff, Timeout: timeout,
		GracePeriod: grace, PollInterval: poll, Env: copyEnv(entry.Env), SourcePath: sourcePath,
	}, nil
}

func parseDuration(raw string, fallback time.Duration) (time.Duration, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fallback, nil
	}
	return time.ParseDuration(raw)
}

func parseOptionalDuration(raw string) (time.Duration, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, nil
	}
	return time.ParseDuration(raw)
}

func copyEnv(in map[string]string) map[string]string {
	if len(in) == 0 {
		return map[string]string{}
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
