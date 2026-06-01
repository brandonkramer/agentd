package managed

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"maps"
	"os"
	"slices"

	"github.com/brandonkramer/catalogfile"
)

// TaskBinding carries work-layer fields included in managed desired hash.
type TaskBinding struct {
	Title      string
	Body       string
	UpdatedAt  string
	HashExtras map[string]any
}

// HashSpec returns a canonical sha256: hash for the resolved agent spec.
func HashSpec(home string, spec *AgentSpec, bind *TaskBinding, store Store) (string, error) {
	payload, err := hashPayload(home, spec, bind, store)
	if err != nil {
		return "", err
	}
	return catalogfile.SHA256Canonical(payload)
}

// SecretsEpoch returns a deterministic hash of secrets/ file contents.
func SecretsEpoch(home string, store Store) (string, error) {
	return catalogfile.DirContentEpoch(store.SecretsDir(home))
}

func hashPayload(home string, spec *AgentSpec, bind *TaskBinding, store Store) (map[string]any, error) {
	envKeys := slices.Sorted(maps.Keys(spec.Env))
	env := make(map[string]string, len(envKeys))
	for _, k := range envKeys {
		env[k] = spec.Env[k]
	}
	promptFileHash := ""
	if spec.PromptFile != "" {
		data, err := os.ReadFile(spec.PromptFile)
		if err != nil {
			return nil, fmt.Errorf("read prompt_file: %w", err)
		}
		sum := sha256.Sum256(data)
		promptFileHash = hex.EncodeToString(sum[:])
	}
	epoch, err := SecretsEpoch(home, store)
	if err != nil {
		return nil, fmt.Errorf("hash spec %q: %w", spec.Name, err)
	}
	out := map[string]any{
		"name": spec.Name, "enabled": spec.Enabled, "task_id": spec.TaskID,
		"title": spec.Title, "body": spec.Body, "driver": spec.Driver, "work_dir": spec.WorkDir,
		"prompt": spec.Prompt, "prompt_file": spec.PromptFile, "prompt_file_hash": promptFileHash,
		"restart": spec.Restart, "max_restarts": spec.MaxRestarts,
		"backoff": spec.Backoff.String(), "timeout": spec.Timeout.String(),
		"grace_period": spec.GracePeriod.String(), "poll_interval": spec.PollInterval.String(),
		"env": env, "secrets_epoch": epoch,
	}
	if bind != nil {
		out["task_title"] = bind.Title
		out["task_body"] = bind.Body
		out["task_updated_at"] = bind.UpdatedAt
		maps.Copy(out, bind.HashExtras)
	}
	return out, nil
}
