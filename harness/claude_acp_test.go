package harness_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/brandonkramer/agentd/harness"
)

func TestClaudeACPHarness(t *testing.T) {
	bin := t.TempDir()
	acp := filepath.Join(bin, "claude-code-acp")
	if err := os.WriteFile(acp, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))

	got, err := harness.Get(harness.ClaudeACP)
	if err != nil {
		t.Fatal(err)
	}
	path, args, err := got.Build(harness.BuildInput{Prompt: "hi", Model: "opus"})
	if err != nil || path != acp {
		t.Fatalf("path=%q args=%v err=%v", path, args, err)
	}
	prep, err := got.Prepare(&harness.WorkInput{RunDir: t.TempDir(), PromptContent: "x", WorkDir: t.TempDir(), Model: "opus"})
	if err != nil {
		t.Fatal(err)
	}
	if prep.Driver != harness.ClaudeACP || prep.Protocol != harness.ProtocolACP || prep.Harness != harness.ClaudeCode {
		t.Fatalf("prep=%+v", prep)
	}
	if len(prep.Warnings) != 1 {
		t.Fatalf("warnings=%v", prep.Warnings)
	}
}

func TestClaudeACPMissingExecutable(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	_, err := harness.FindClaudeACPCLI()
	if err == nil {
		t.Fatal("expected error")
	}
}
