package agentd_test

import (
	"testing"

	"github.com/brandonkramer/agentd"
)

func TestDefaultLayout(t *testing.T) {
	t.Parallel()

	if agentd.Default.PipePrefix != "agentd" {
		t.Fatalf("pipe prefix=%q", agentd.Default.PipePrefix)
	}
	if agentd.Agentd.EnvVar != "AGENTD_HOME" || agentd.Agentd.DefaultDir != ".agentd" {
		t.Fatalf("service=%+v", agentd.Agentd)
	}
}
