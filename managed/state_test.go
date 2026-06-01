package managed

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStateRoundTrip(t *testing.T) {
	home := t.TempDir()
	st := &State{Name: "bot", DesiredHash: "d", AppliedHash: "a", RunID: "r1"}
	if err := SaveState(home, st, DefaultStore); err != nil {
		t.Fatal(err)
	}
	got, err := LoadState(home, "bot", DefaultStore)
	if err != nil || got.Name != "bot" || got.RunID != "r1" {
		t.Fatalf("LoadState=%+v err=%v", got, err)
	}
	states, err := ListStates(home, DefaultStore)
	if err != nil || len(states) != 1 {
		t.Fatalf("ListStates=%+v err=%v", states, err)
	}
	if err := DeleteState(home, "bot", DefaultStore); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadState(home, "bot", DefaultStore); err == nil {
		t.Fatal("expected missing state")
	}
	if err := DeleteState(home, "bot", DefaultStore); err != nil {
		t.Fatal(err)
	}
}

func TestStatePathValidation(t *testing.T) {
	if _, err := statePath("/tmp", "..", DefaultStore); err == nil {
		t.Fatal("expected invalid name")
	}
	if _, err := statePath("/tmp", "", DefaultStore); err == nil {
		t.Fatal("expected empty name error")
	}
}

func TestListStatesMissingDir(t *testing.T) {
	got, err := ListStates(t.TempDir(), DefaultStore)
	if err != nil || len(got) != 0 {
		t.Fatalf("got=%+v err=%v", got, err)
	}
}

func TestListStatesSkipsBadEntries(t *testing.T) {
	home := t.TempDir()
	dir := DefaultStore.StateDir(home)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	_ = os.WriteFile(filepath.Join(dir, "sub", ".json"), []byte("{}"), 0o644)
	if err := os.WriteFile(filepath.Join(dir, "bad.json"), []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "dir"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := SaveState(home, &State{Name: "ok"}, DefaultStore); err != nil {
		t.Fatal(err)
	}
	states, err := ListStates(home, DefaultStore)
	if err != nil || len(states) != 1 || states[0].Name != "ok" {
		t.Fatalf("states=%+v err=%v", states, err)
	}
}
