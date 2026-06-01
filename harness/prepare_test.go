package harness_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/brandonkramer/agentd/harness"
)

func TestGenericPrepareAndSubstitute(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	in := &harness.WorkInput{
		RunDir:          dir,
		CommandTemplate: "cat {prompt}",
		PromptContent:   "hello",
		Harness:         harness.GenericCommand,
	}
	got, err := harness.PrepareGeneric(in)
	if err != nil {
		t.Fatal(err)
	}
	if got.Command == in.CommandTemplate {
		t.Fatalf("expected substitution, got %q", got.Command)
	}
	if got.PromptPath == "" {
		t.Fatal("expected prompt path")
	}
	if _, err := harness.PrepareGeneric(&harness.WorkInput{RunDir: dir}); err == nil {
		t.Fatal("expected missing command error")
	}
}

func TestPrepareGenericPromptWriteError(t *testing.T) {
	t.Parallel()

	in := &harness.WorkInput{
		RunDir:          "/",
		CommandTemplate: "echo hi",
		PromptContent:   "x",
	}
	if _, err := harness.PrepareGeneric(in); err == nil {
		t.Fatal("expected write error")
	}
}

func TestClaudePrepare(t *testing.T) {
	bin := t.TempDir()
	claude := filepath.Join(bin, "claude")
	if err := os.WriteFile(claude, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))

	dir := t.TempDir()
	got, err := harness.PrepareClaude(&harness.WorkInput{
		RunDir: dir, PromptContent: "do work", WorkDir: dir, Model: "opus",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.ExecPath != claude || !got.StdinPrompt || got.PromptPath == "" {
		t.Fatalf("got=%+v", got)
	}
	if harness.ClaudeBaseArgs()[0] != "-p" {
		t.Fatal("unexpected base args")
	}
}

func TestFindClaudeCLIMissing(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	if _, err := harness.FindClaudeCLI(); err == nil {
		t.Fatal("expected error")
	}
}

func TestClaudePrepareErrors(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	if _, err := harness.PrepareClaude(&harness.WorkInput{RunDir: t.TempDir(), PromptContent: "x"}); err == nil {
		t.Fatal("expected missing claude")
	}
}

func TestClaudePrepareWithFakeBinary(t *testing.T) {
	bin := t.TempDir()
	claude := filepath.Join(bin, "claude")
	if err := os.WriteFile(claude, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin)
	got, err := harness.PrepareClaude(&harness.WorkInput{
		RunDir: t.TempDir(), PromptContent: "x", WorkDir: t.TempDir(), Model: "m",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Driver != harness.ClaudeCode || got.ExecPath != claude {
		t.Fatalf("got=%+v", got)
	}
}
