package harness_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/brandonkramer/agentd/harness"
)

func TestBuiltinHarnessesRegistered(t *testing.T) {
	g, err := harness.Get(harness.GenericCommand)
	if err != nil {
		t.Fatal(err)
	}
	bin, args, err := g.Build(harness.BuildInput{Prompt: "echo hi"})
	if err != nil || bin != "sh" || len(args) != 2 {
		t.Fatalf("build=%q %v err=%v", bin, args, err)
	}
	if _, _, err := g.Build(harness.BuildInput{}); err == nil {
		t.Fatal("expected empty prompt error")
	}
	prep, err := g.Prepare(&harness.WorkInput{
		RunDir: t.TempDir(), CommandTemplate: "echo hi", PromptContent: "x",
	})
	if err != nil {
		t.Fatal(err)
	}
	if prep.Driver != harness.GenericCommand {
		t.Fatalf("prep=%+v", prep)
	}
}

func TestErrRequired(t *testing.T) {
	g, err := harness.Get(harness.GenericCommand)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := g.Build(harness.BuildInput{}); err == nil || err.Error() == "" {
		t.Fatalf("err=%v", err)
	}
}

func TestClaudeHarness(t *testing.T) {
	bin := t.TempDir()
	claude := filepath.Join(bin, "claude")
	if err := os.WriteFile(claude, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	g, err := harness.Get(harness.ClaudeCode)
	if err != nil {
		t.Fatal(err)
	}
	path, args, err := g.Build(harness.BuildInput{Prompt: "hi", Model: "m"})
	if err != nil || path != claude {
		t.Fatalf("path=%q args=%v err=%v", path, args, err)
	}
	prep, err := g.Prepare(&harness.WorkInput{RunDir: t.TempDir(), PromptContent: "x", WorkDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	if prep.Driver != harness.ClaudeCode {
		t.Fatalf("prep=%+v", prep)
	}
}
