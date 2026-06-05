package harness_test

import (
	"testing"

	"github.com/brandonkramer/agentd/harness"
)

func TestResolveDriverExplicit(t *testing.T) {
	t.Parallel()
	name, err := harness.ResolveDriver(harness.ClaudeCode, "", "")
	if err != nil || name != harness.ClaudeCode {
		t.Fatalf("name=%q err=%v", name, err)
	}
}

func TestResolveDriverDefaultsToGenericCommand(t *testing.T) {
	t.Parallel()
	name, err := harness.ResolveDriver("", "", "echo hi")
	if err != nil || name != harness.GenericCommand {
		t.Fatalf("name=%q err=%v", name, err)
	}
}

func TestResolveDriverUsesAgentDriver(t *testing.T) {
	t.Parallel()
	name, err := harness.ResolveDriver("", harness.ClaudeCode, "")
	if err != nil || name != harness.ClaudeCode {
		t.Fatalf("name=%q err=%v", name, err)
	}
}

func TestResolveDriverRequired(t *testing.T) {
	t.Parallel()
	if _, err := harness.ResolveDriver("", "", ""); err == nil {
		t.Fatal("expected error")
	}
}
