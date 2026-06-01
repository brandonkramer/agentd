package managed

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/brandonkramer/jsonfile"
)

// State is observed reconciler state for one managed agent.
type State struct {
	Name          string `json:"name"`
	DesiredHash   string `json:"desired_hash,omitempty"`
	AppliedHash   string `json:"applied_hash,omitempty"`
	RunID         string `json:"run_id,omitempty"`
	Restarts      int    `json:"restarts"`
	LastExitCode  *int   `json:"last_exit_code"`
	LastError     string `json:"last_error,omitempty"`
	LastStartedAt string `json:"last_started_at,omitempty"`
	Converging    bool   `json:"converging"`
	TaskID        string `json:"task_id,omitempty"`
}

// LoadState reads persisted state for one managed agent.
func LoadState(home, name string, store Store) (State, error) {
	path, err := statePath(home, name, store)
	if err != nil {
		return State{}, fmt.Errorf("load state %q: %w", name, err)
	}
	st, err := jsonfile.Read[State](path)
	if err != nil {
		return State{}, fmt.Errorf("load state %q: read %s: %w", name, path, err)
	}
	return st, nil
}

// SaveState persists observed state for one managed agent.
func SaveState(home string, st *State, store Store) error {
	path, err := statePath(home, st.Name, store)
	if err != nil {
		return fmt.Errorf("save state %q: %w", st.Name, err)
	}
	if err := jsonfile.Write(path, st); err != nil {
		return fmt.Errorf("save state %q: write %s: %w", st.Name, path, err)
	}
	return nil
}

// DeleteState removes persisted state for one managed agent.
func DeleteState(home, name string, store Store) error {
	path, err := statePath(home, name, store)
	if err != nil {
		return fmt.Errorf("delete state %q: %w", name, err)
	}
	err = os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("delete state %q: remove %s: %w", name, path, err)
	}
	return nil
}

// ListStates returns all persisted managed-agent states under home.
func ListStates(home string, store Store) ([]State, error) {
	dir := store.StateDir(home)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("list states: read %s: %w", dir, err)
	}
	out := make([]State, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		name := e.Name()[:len(e.Name())-len(".json")]
		st, err := LoadState(home, name, store)
		if err != nil {
			continue
		}
		out = append(out, st)
	}
	return out, nil
}

func statePath(home, name string, store Store) (string, error) {
	if name == "" || name != filepath.Base(name) || name == "." || name == ".." {
		return "", os.ErrInvalid
	}
	return filepath.Join(store.StateDir(home), name+".json"), nil
}
