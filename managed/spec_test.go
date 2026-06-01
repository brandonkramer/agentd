package managed

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMergedSpecsProjectOverride(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	global := filepath.Join(home, "agents.toml")
	if err := os.WriteFile(global, []byte(`[[agents]]
name = "bot"
enabled = true
driver = "generic-command"
work_dir = "/tmp/global"
prompt = "echo global"
`), 0o644); err != nil {
		t.Fatal(err)
	}
	projectFile := filepath.Join(project, ".agentd", "agents.toml")
	if err := os.MkdirAll(filepath.Dir(projectFile), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(projectFile, []byte(`[[agents]]
name = "bot"
enabled = true
driver = "generic-command"
prompt = "echo project"
`), 0o644); err != nil {
		t.Fatal(err)
	}
	specs, err := LoadMergedSpecs(home, project, DefaultStore)
	if err != nil {
		t.Fatal(err)
	}
	if specs["bot"].Prompt != "echo project" {
		t.Fatalf("prompt=%q", specs["bot"].Prompt)
	}
	if specs["bot"].WorkDir != project {
		t.Fatalf("work_dir=%q", specs["bot"].WorkDir)
	}
}
