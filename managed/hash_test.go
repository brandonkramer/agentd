package managed

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHashSpecStable(t *testing.T) {
	home := t.TempDir()
	spec := &AgentSpec{
		Name: "bot", Enabled: true, Driver: "generic-command", WorkDir: "/tmp",
		Prompt: "echo hi", Restart: "no",
	}
	h1, err := HashSpec(home, spec, nil, DefaultStore)
	if err != nil {
		t.Fatal(err)
	}
	h2, err := HashSpec(home, spec, nil, DefaultStore)
	if err != nil {
		t.Fatal(err)
	}
	if h1 != h2 {
		t.Fatalf("hashes differ: %q vs %q", h1, h2)
	}
}

func TestSecretsEpochChangesOnFileChange(t *testing.T) {
	home := t.TempDir()
	secrets := filepath.Join(home, "secrets")
	if err := os.MkdirAll(secrets, 0o755); err != nil {
		t.Fatal(err)
	}
	e1, err := SecretsEpoch(home, DefaultStore)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(secrets, "a.json"), []byte("1"), 0o644); err != nil {
		t.Fatal(err)
	}
	e2, err := SecretsEpoch(home, DefaultStore)
	if err != nil {
		t.Fatal(err)
	}
	if e1 == e2 {
		t.Fatal("expected secrets epoch to change")
	}
}

func TestHashSpecIncludesTask(t *testing.T) {
	home := t.TempDir()
	spec := &AgentSpec{Name: "bot", Enabled: true, Driver: "generic-command", WorkDir: "/tmp", Prompt: "x"}
	task := &TaskBinding{Title: "t", Body: "b", UpdatedAt: "2026-01-01T00:00:00Z"}
	h1, _ := HashSpec(home, spec, task, DefaultStore)
	task.UpdatedAt = "2026-01-02T00:00:00Z"
	h2, _ := HashSpec(home, spec, task, DefaultStore)
	if h1 == h2 {
		t.Fatal("expected task update to change hash")
	}
}

func TestHashSpecPromptFileAndErrors(t *testing.T) {
	home := t.TempDir()
	prompt := filepath.Join(home, "p.txt")
	if err := os.WriteFile(prompt, []byte("body"), 0o644); err != nil {
		t.Fatal(err)
	}
	spec := &AgentSpec{
		Name: "bot", Enabled: true, Driver: "generic-command", WorkDir: "/tmp",
		PromptFile: prompt, Restart: "no", Prompt: "fallback",
	}
	if _, err := HashSpec(home, spec, nil, DefaultStore); err != nil {
		t.Fatal(err)
	}
	spec.PromptFile = filepath.Join(home, "missing.txt")
	if _, err := HashSpec(home, spec, nil, DefaultStore); err == nil {
		t.Fatal("expected prompt file error")
	}
}

func TestHashSpecHashExtras(t *testing.T) {
	home := t.TempDir()
	spec := &AgentSpec{Name: "bot", Enabled: true, Driver: "generic-command", WorkDir: "/tmp", Prompt: "x"}
	bind := &TaskBinding{HashExtras: map[string]any{"extra": 1}}
	if _, err := HashSpec(home, spec, bind, DefaultStore); err != nil {
		t.Fatal(err)
	}
}
