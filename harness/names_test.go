package harness

import "testing"

func TestHarnessNames(t *testing.T) {
	t.Parallel()

	if (genericCommand{}).Name() != GenericCommand {
		t.Fatal("generic name")
	}
	if (claudeCode{}).Name() != ClaudeCode {
		t.Fatal("claude name")
	}
	if (piDriver{}).Name() != Pi {
		t.Fatal("pi name")
	}
}
