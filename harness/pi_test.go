package harness_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/brandonkramer/agentd/harness"
)

func TestPiHarnessRegistered(t *testing.T) {
	if _, err := harness.Get(harness.Pi); err != nil {
		t.Fatal(err)
	}
}

func TestPiHarnessPrepare(t *testing.T) {
	binDir := t.TempDir()
	piBin := filepath.Join(binDir, "pi")
	if runtime.GOOS == "windows" {
		piBin += ".exe"
	}
	if err := os.WriteFile(piBin, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	prep, err := harness.PreparePi(&harness.WorkInput{
		RunDir: t.TempDir(), WorkDir: t.TempDir(), PromptContent: "ping",
	})
	if err != nil {
		t.Fatal(err)
	}
	if prep.Driver != harness.Pi || prep.Protocol != harness.ProtocolPiJSON {
		t.Fatalf("prep=%+v", prep)
	}
}
