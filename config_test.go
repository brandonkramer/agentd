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

func TestWorkService(t *testing.T) {
	t.Parallel()

	if agentd.WorkLayout.PipePrefix != "agentd-work" {
		t.Fatalf("work pipe prefix=%q", agentd.WorkLayout.PipePrefix)
	}
	if agentd.Work.EnvVar != "AGENTD_HOME" || agentd.Work.DefaultDir != ".agentd-work" {
		t.Fatalf("work service=%+v", agentd.Work)
	}
	if agentd.Work.RegistryFile != "sessions/known-daemons.json" {
		t.Fatalf("registry file=%q", agentd.Work.RegistryFile)
	}
}
