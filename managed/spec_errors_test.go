package managed

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadMergedSpecsErrors(t *testing.T) {
	home := t.TempDir()
	bad := filepath.Join(home, "agents.toml")
	if err := os.WriteFile(bad, []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadMergedSpecs(home, "", DefaultStore); err == nil {
		t.Fatal("expected parse error")
	}
}

func TestNormalizeEntryValidation(t *testing.T) {
	cases := []struct {
		name string
		body string
	}{
		{name: "missing name", body: `[[agents]]\nprompt="x"`},
		{name: "invalid restart", body: `[[agents]]\nname="a"\nrestart="maybe"\nprompt="x"`},
		{name: "bad backoff", body: `[[agents]]\nname="a"\nbackoff="nope"\nprompt="x"`},
		{name: "bad timeout", body: `[[agents]]\nname="a"\ntimeout="nope"\nprompt="x"`},
		{name: "bad grace", body: `[[agents]]\nname="a"\ngrace_period="nope"\nprompt="x"`},
		{name: "bad poll", body: `[[agents]]\nname="a"\npoll_interval="nope"\nprompt="x"`},
		{name: "unknown driver", body: `[[agents]]\nname="a"\ndriver="nope"\nprompt="x"`},
		{name: "task conflict", body: `[[agents]]\nname="a"\ntask_id="1"\ntitle="t"\nprompt="x"`},
		{name: "no work", body: `[[agents]]\nname="a"`},
		{name: "duplicate", body: `[[agents]]\nname="a"\nprompt="x"\n\n[[agents]]\nname="a"\nprompt="y"`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "agents.toml")
			if err := os.WriteFile(path, []byte(tc.body), 0o644); err != nil {
				t.Fatal(err)
			}
			if _, err := loadFile(path, t.TempDir(), ""); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestCopyEnv(t *testing.T) {
	out := copyEnv(map[string]string{"A": "1"})
	if out["A"] != "1" {
		t.Fatalf("out=%v", out)
	}
	if len(copyEnv(nil)) != 0 {
		t.Fatal("expected empty map")
	}
}

func TestParseDurations(t *testing.T) {
	if d, err := parseDuration("", time.Second); err != nil || d != time.Second {
		t.Fatalf("d=%v err=%v", d, err)
	}
	if _, err := parseDuration("bad", 0); err == nil {
		t.Fatal("expected error")
	}
	if d, err := parseOptionalDuration(""); err != nil || d != 0 {
		t.Fatalf("d=%v err=%v", d, err)
	}
	if _, err := parseOptionalDuration("bad"); err == nil {
		t.Fatal("expected error")
	}
}

func TestNormalizeEntrySuccessVariants(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "agents.toml")
	body := `[[agents]]
name = "a"
enabled = false
task_id = "t1"
driver = "generic-command"
prompt_file = "p.md"
timeout = "1s"
`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "p.md"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	specs, err := loadFile(path, dir, dir)
	if err != nil {
		t.Fatal(err)
	}
	if !specs["a"].Enabled == false || specs["a"].TaskID != "t1" || specs["a"].Timeout != time.Second {
		t.Fatalf("spec=%+v", specs["a"])
	}
}

func TestLoadMergedSpecsProjectError(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	global := `[[agents]]
name = "a"
prompt = "x"
driver = "generic-command"
`
	if err := os.WriteFile(filepath.Join(home, "agents.toml"), []byte(global), 0o644); err != nil {
		t.Fatal(err)
	}
	bad := filepath.Join(project, ".agentd", "agents.toml")
	if err := os.MkdirAll(filepath.Dir(bad), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(bad, []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadMergedSpecs(home, project, DefaultStore); err == nil {
		t.Fatal("expected project parse error")
	}
}

func TestLoadFileMissing(t *testing.T) {
	got, err := loadFile(filepath.Join(t.TempDir(), "missing.toml"), t.TempDir(), "")
	if err != nil || got != nil {
		t.Fatalf("got=%v err=%v", got, err)
	}
}

func TestNormalizeEntryWorkDirAbs(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "agents.toml")
	body := `[[agents]]
name = "a"
prompt = "x"
driver = "generic-command"
work_dir = "."
`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	specs, err := loadFile(path, dir, dir)
	if err != nil || specs["a"].WorkDir == "" {
		t.Fatalf("specs=%+v err=%v", specs, err)
	}
}

func TestStateLoadError(t *testing.T) {
	home := t.TempDir()
	dir := DefaultStore.StateDir(home)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "bot.json"), []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadState(home, "bot", DefaultStore); err == nil {
		t.Fatal("expected load error")
	}
}

func TestListStatesReadDirError(t *testing.T) {
	home := t.TempDir()
	if err := os.WriteFile(DefaultStore.StateDir(home), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ListStates(home, DefaultStore); err == nil {
		t.Fatal("expected readdir error")
	}
}
