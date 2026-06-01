package managed_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/brandonkramer/agentd/managed"
)

func TestMain(m *testing.M) {
	managed.SetDriverValidator(func(name string) error {
		switch name {
		case "generic-command", "claude-code":
			return nil
		default:
			return fmt.Errorf("unknown harness %q", name)
		}
	})
	os.Exit(m.Run())
}
