package agentd_test

import (
	"context"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/brandonkramer/agentd"
)

func TestChildEnv(t *testing.T) {
	t.Parallel()

	env := agentd.ChildEnv("/home", "run-1", map[string]string{"K": "V"})
	foundRun, foundHome, foundK := false, false, false
	for _, e := range env {
		switch e {
		case agentd.EnvRunID + "=run-1":
			foundRun = true
		case agentd.EnvHome + "=/home":
			foundHome = true
		case "K=V":
			foundK = true
		}
	}
	if !foundRun || !foundHome || !foundK {
		t.Fatalf("env=%v", env)
	}
}

func TestVerifyRunIDBasics(t *testing.T) {
	t.Parallel()

	if agentd.VerifyRunID("x", 0, nil, nil) {
		t.Fatal("pid 0 should fail")
	}
	if agentd.VerifyRunID("x", os.Getpid(), func(int) bool { return false }, nil) {
		t.Fatal("not alive should fail")
	}
}

func TestVerifyRunIDSubprocess(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "sleep", "2")
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = cmd.Process.Kill() }()
	if agentd.VerifyRunID("ignored", cmd.Process.Pid, func(int) bool { return true }, nil) {
		t.Fatal("expected no env marker match")
	}
}
