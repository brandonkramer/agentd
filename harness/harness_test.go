package harness_test

import (
	"testing"

	"github.com/brandonkramer/agentd/harness"
)

func TestGlobalRegistry(t *testing.T) {
	if _, err := harness.Get("missing"); err == nil {
		t.Fatal("expected unknown harness")
	}
	names := harness.Names()
	if len(names) == 0 {
		t.Fatal("expected builtins")
	}
	exes := harness.ExecutableNames()
	if len(exes) == 0 {
		t.Fatal("expected executables")
	}
}
