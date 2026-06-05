package harness_test

import (
	"testing"

	"github.com/brandonkramer/agentd/harness"
)

func TestRegistryKnownNames(t *testing.T) {
	names := harness.Names()
	if len(names) < 2 {
		t.Fatalf("names=%v", names)
	}
	for _, name := range []string{"generic-command", "claude-code", "pi"} {
		if _, err := harness.Get(name); err != nil {
			t.Fatalf("Get(%q): %v", name, err)
		}
	}
}

func TestRegistryUnknown(t *testing.T) {
	if _, err := harness.Get("no-such-harness"); err == nil {
		t.Fatal("expected error")
	}
}

func TestExecutableNamesNonEmpty(t *testing.T) {
	if len(harness.ExecutableNames()) == 0 {
		t.Fatal("expected executable names")
	}
}
