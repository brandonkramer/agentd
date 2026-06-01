package managed_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/brandonkramer/reconciler"

	"github.com/brandonkramer/agentd/managed"
)

type fakeEngine struct {
	mu           sync.Mutex
	works        map[string]reconciler.WorkStatus
	starts       int
	stops        []string
	shuttingDown bool
	bindErr      error
	startErr     error
	now          string
}

func (e *fakeEngine) StartWork(_ context.Context, _ *managed.AgentSpec, _ *managed.State) (string, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.startErr != nil {
		return "", e.startErr
	}
	e.starts++
	id := "work-1"
	e.works[id] = reconciler.WorkStatus{Terminal: false}
	go func() {
		time.Sleep(20 * time.Millisecond)
		e.mu.Lock()
		e.works[id] = reconciler.WorkStatus{Terminal: true}
		e.mu.Unlock()
	}()
	return id, nil
}

func (e *fakeEngine) ResolveBinding(_ *managed.AgentSpec, _ *managed.State) (*managed.TaskBinding, error) {
	if e.bindErr != nil {
		return nil, e.bindErr
	}
	return &managed.TaskBinding{Title: "t", Body: "b"}, nil
}

func (e *fakeEngine) LoadWork(workID string) (reconciler.WorkStatus, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	st, ok := e.works[workID]
	if !ok {
		return reconciler.WorkStatus{}, os.ErrNotExist
	}
	return st, nil
}

func (e *fakeEngine) GracefulStopWork(workID string, _ time.Duration) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.stops = append(e.stops, workID)
}

func (e *fakeEngine) ShuttingDown() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.shuttingDown
}

func (e *fakeEngine) NowISO() string {
	if e.now != "" {
		return e.now
	}
	return "2026-01-01T00:00:00Z"
}

func (e *fakeEngine) StartCount() int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.starts
}

func (e *fakeEngine) StopCount() int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return len(e.stops)
}

func waitWorkDone(t *testing.T, home, name string) {
	t.Helper()
	waitFor(t, func() bool {
		st, _ := managed.LoadState(home, name, managed.DefaultStore)
		return st.AppliedHash != "" && st.LastError == "" && st.RunID == ""
	})
}

func stopManager(t *testing.T, mgr *reconciler.Manager) {
	t.Helper()
	mgr.StopAll()
	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
}

func writeAgent(t *testing.T, home, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(home, "agents.toml"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func waitFor(t *testing.T, fn func() bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if fn() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("condition not met")
}

func TestManagerTickStartsAndConverges(t *testing.T) {
	home := t.TempDir()
	writeAgent(t, home, `[[agents]]
name = "bot"
enabled = true
driver = "generic-command"
work_dir = "/tmp"
prompt = "echo hi"
restart = "no"
poll_interval = "10ms"
`)
	eng := &fakeEngine{works: map[string]reconciler.WorkStatus{}}
	mgr, err := managed.NewReconcilerManager(home, "", eng, managed.DefaultStore)
	if err != nil {
		t.Fatal(err)
	}
	if err := mgr.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	waitWorkDone(t, home, "bot")
	if eng.StartCount() != 1 {
		t.Fatalf("starts=%d", eng.StartCount())
	}
}

func TestManagerTickDisabledAgentClearsState(t *testing.T) {
	home := t.TempDir()
	writeAgent(t, home, `[[agents]]
name = "bot"
enabled = false
driver = "generic-command"
prompt = "x"
`)
	if err := managed.SaveState(home, &managed.State{Name: "bot", LastError: "old"}, managed.DefaultStore); err != nil {
		t.Fatal(err)
	}
	mgr, err := managed.NewReconcilerManager(home, "", &fakeEngine{works: map[string]reconciler.WorkStatus{}}, managed.DefaultStore)
	if err != nil {
		t.Fatal(err)
	}
	if err := mgr.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	st, _ := managed.LoadState(home, "bot", managed.DefaultStore)
	if st.Name != "" {
		t.Fatalf("state=%+v", st)
	}
}

func TestManagerTickRecordsBindingError(t *testing.T) {
	home := t.TempDir()
	writeAgent(t, home, `[[agents]]
name = "bot"
enabled = true
driver = "generic-command"
prompt = "x"
`)
	eng := &fakeEngine{works: map[string]reconciler.WorkStatus{}, bindErr: os.ErrInvalid}
	mgr, err := managed.NewReconcilerManager(home, "", eng, managed.DefaultStore)
	if err != nil {
		t.Fatal(err)
	}
	if err := mgr.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	st, _ := managed.LoadState(home, "bot", managed.DefaultStore)
	if st.LastError == "" {
		t.Fatal("expected error recorded")
	}
}

func TestManagerTickRestartOnHashChange(t *testing.T) {
	home := t.TempDir()
	writeAgent(t, home, `[[agents]]
name = "bot"
enabled = true
driver = "generic-command"
work_dir = "/tmp"
prompt = "echo one"
restart = "no"
poll_interval = "10ms"
`)
	eng := &fakeEngine{works: map[string]reconciler.WorkStatus{}}
	mgr, err := managed.NewReconcilerManager(home, "", eng, managed.DefaultStore)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := mgr.Tick(ctx); err != nil {
		t.Fatal(err)
	}
	waitFor(t, func() bool {
		st, _ := managed.LoadState(home, "bot", managed.DefaultStore)
		return st.AppliedHash != ""
	})

	writeAgent(t, home, `[[agents]]
name = "bot"
enabled = true
driver = "generic-command"
work_dir = "/tmp"
prompt = "echo two"
restart = "no"
poll_interval = "10ms"
`)
	if err := mgr.Tick(ctx); err != nil {
		t.Fatal(err)
	}
	waitFor(t, func() bool { return eng.StopCount() > 0 })
	stopManager(t, mgr)
}

func TestManagerStopAll(t *testing.T) {
	home := t.TempDir()
	writeAgent(t, home, `[[agents]]
name = "bot"
enabled = true
driver = "generic-command"
work_dir = "/tmp"
prompt = "echo hi"
restart = "no"
poll_interval = "10ms"
`)
	mgr, err := managed.NewReconcilerManager(home, "", &fakeEngine{works: map[string]reconciler.WorkStatus{}}, managed.DefaultStore)
	if err != nil {
		t.Fatal(err)
	}
	if err := mgr.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	waitFor(t, func() bool {
		st, _ := managed.LoadState(home, "bot", managed.DefaultStore)
		return st.AppliedHash != ""
	})
	stopManager(t, mgr)
}

func TestManagerRunLoopStopsOnCancel(t *testing.T) {
	home := t.TempDir()
	writeAgent(t, home, "")
	mgr, err := managed.NewReconcilerManager(home, "", &fakeEngine{works: map[string]reconciler.WorkStatus{}}, managed.DefaultStore)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		mgr.RunLoop(ctx)
		close(done)
	}()
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("RunLoop did not stop")
	}
}

func TestShouldRestartPolicies(t *testing.T) {
	t.Parallel()

	ok := 0
	fail := 1
	cases := []struct {
		name       string
		policy     string
		failed     bool
		exitCode   *int
		userCancel bool
		want       bool
	}{
		{name: "no ignores failure", policy: "no", failed: true, want: false},
		{name: "on-failure failed run", policy: "on-failure", failed: true, want: true},
		{name: "on-failure success exit", policy: "on-failure", exitCode: &ok, want: false},
		{name: "on-failure nonzero exit", policy: "on-failure", exitCode: &fail, want: true},
		{name: "always restarts", policy: "always", exitCode: &ok, want: true},
		{name: "user cancel blocks always", policy: "always", exitCode: &ok, userCancel: true, want: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := reconciler.ParseRestartPolicy(tc.policy).ShouldRestart(tc.failed, tc.exitCode, tc.userCancel)
			if got != tc.want {
				t.Fatalf("ShouldRestart(%q, %v, %v, %v) = %v; want %v",
					tc.policy, tc.failed, tc.exitCode, tc.userCancel, got, tc.want)
			}
		})
	}
}

func TestShouldRestartOnFailureExit(t *testing.T) {
	code := 2
	if !reconciler.RestartOnFailure.ShouldRestart(false, &code, false) {
		t.Fatal("expected restart on nonzero exit")
	}
}

type richEngine struct {
	fakeEngine
	failLoad   bool
	exitCode   *int
	failed     bool
	userCancel bool
}

func (e *richEngine) LoadWork(workID string) (reconciler.WorkStatus, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.failLoad {
		return reconciler.WorkStatus{}, os.ErrNotExist
	}
	st, ok := e.works[workID]
	if !ok {
		return reconciler.WorkStatus{}, os.ErrNotExist
	}
	if st.Terminal {
		st.ExitCode = e.exitCode
		st.Failed = e.failed
		st.UserCancel = e.userCancel
	}
	return st, nil
}

func TestManagerSuperviseRestartAndErrors(t *testing.T) {
	home := t.TempDir()
	writeAgent(t, home, `[[agents]]
name = "bot"
enabled = true
driver = "generic-command"
work_dir = "/tmp"
prompt = "echo hi"
restart = "on-failure"
max_restarts = 1
backoff = "5ms"
poll_interval = "5ms"
`)
	eng := &richEngine{fakeEngine: fakeEngine{works: map[string]reconciler.WorkStatus{}}, failed: true}
	eng.works["work-1"] = reconciler.WorkStatus{Terminal: true, Failed: true}
	mgr, err := managed.NewReconcilerManager(home, "", eng, managed.DefaultStore)
	if err != nil {
		t.Fatal(err)
	}
	ctx := t.Context()
	if err := mgr.Tick(ctx); err != nil {
		t.Fatal(err)
	}
	waitFor(t, func() bool { return eng.StartCount() >= 1 })
	time.Sleep(80 * time.Millisecond)
	stopManager(t, mgr)
}

func TestManagerStartWorkError(t *testing.T) {
	home := t.TempDir()
	writeAgent(t, home, `[[agents]]
name = "bot"
enabled = true
driver = "generic-command"
prompt = "x"
poll_interval = "5ms"
`)
	eng := &fakeEngine{works: map[string]reconciler.WorkStatus{}, startErr: errors.New("start")}
	mgr, err := managed.NewReconcilerManager(home, "", eng, managed.DefaultStore)
	if err != nil {
		t.Fatal(err)
	}
	if err := mgr.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	time.Sleep(50 * time.Millisecond)
	stopManager(t, mgr)
	st, _ := managed.LoadState(home, "bot", managed.DefaultStore)
	if st.LastError == "" {
		t.Fatal("expected start error")
	}
}

func TestManagerTickRemovesAgent(t *testing.T) {
	home := t.TempDir()
	writeAgent(t, home, `[[agents]]
name = "bot"
enabled = true
driver = "generic-command"
prompt = "x"
poll_interval = "5ms"
`)
	mgr, err := managed.NewReconcilerManager(home, "", &fakeEngine{works: map[string]reconciler.WorkStatus{}}, managed.DefaultStore)
	if err != nil {
		t.Fatal(err)
	}
	if err := mgr.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	writeAgent(t, home, "")
	if err := mgr.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	stopManager(t, mgr)
}

func TestManagerHashSpecError(t *testing.T) {
	home := t.TempDir()
	writeAgent(t, home, `[[agents]]
name = "bot"
enabled = true
driver = "generic-command"
prompt_file = "missing.md"
prompt = "x"
`)
	mgr, err := managed.NewReconcilerManager(home, "", &fakeEngine{works: map[string]reconciler.WorkStatus{}}, managed.DefaultStore)
	if err != nil {
		t.Fatal(err)
	}
	if err := mgr.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	st, _ := managed.LoadState(home, "bot", managed.DefaultStore)
	if st.LastError == "" {
		t.Fatal("expected hash error")
	}
}

func TestManagerRunLoopTicks(t *testing.T) {
	home := t.TempDir()
	writeAgent(t, home, "")
	mgr, err := managed.NewReconcilerManager(home, "", &fakeEngine{works: map[string]reconciler.WorkStatus{}}, managed.DefaultStore)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()
	mgr.RunLoop(ctx)
}

func TestShouldRestartDefault(t *testing.T) {
	if reconciler.ParseRestartPolicy("weird").ShouldRestart(false, nil, false) {
		t.Fatal("default false")
	}
}

func TestManagerExistingRunningWork(t *testing.T) {
	home := t.TempDir()
	writeAgent(t, home, `[[agents]]
name = "bot"
enabled = true
driver = "generic-command"
prompt = "x"
poll_interval = "5ms"
`)
	eng := &fakeEngine{works: map[string]reconciler.WorkStatus{"work-1": {Terminal: false}}}
	if err := managed.SaveState(home, &managed.State{Name: "bot", RunID: "work-1", AppliedHash: "same"}, managed.DefaultStore); err != nil {
		t.Fatal(err)
	}
	mgr, err := managed.NewReconcilerManager(home, "", eng, managed.DefaultStore)
	if err != nil {
		t.Fatal(err)
	}
	if err := mgr.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	time.Sleep(30 * time.Millisecond)
	stopManager(t, mgr)
}

func TestManagerSuperviseHashChangeExit(t *testing.T) {
	home := t.TempDir()
	writeAgent(t, home, `[[agents]]
name = "bot"
enabled = true
driver = "generic-command"
prompt = "one"
poll_interval = "5ms"
`)
	eng := &fakeEngine{works: map[string]reconciler.WorkStatus{}}
	mgr, err := managed.NewReconcilerManager(home, "", eng, managed.DefaultStore)
	if err != nil {
		t.Fatal(err)
	}
	if err := mgr.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	waitFor(t, func() bool {
		st, _ := managed.LoadState(home, "bot", managed.DefaultStore)
		return st.AppliedHash != ""
	})
	writeAgent(t, home, `[[agents]]
name = "bot"
enabled = true
driver = "generic-command"
prompt = "two"
poll_interval = "5ms"
`)
	if err := mgr.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	stopManager(t, mgr)
}

func TestManagerMaxRestarts(t *testing.T) {
	home := t.TempDir()
	writeAgent(t, home, `[[agents]]
name = "bot"
enabled = true
driver = "generic-command"
prompt = "x"
restart = "always"
max_restarts = 0
poll_interval = "5ms"
backoff = "1ms"
`)
	eng := &fakeEngine{works: map[string]reconciler.WorkStatus{}}
	mgr, err := managed.NewReconcilerManager(home, "", eng, managed.DefaultStore)
	if err != nil {
		t.Fatal(err)
	}
	if err := mgr.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	time.Sleep(80 * time.Millisecond)
	stopManager(t, mgr)
}

func TestWaitUntilTerminalLoadError(t *testing.T) {
	home := t.TempDir()
	writeAgent(t, home, `[[agents]]
name = "bot"
enabled = true
driver = "generic-command"
prompt = "x"
poll_interval = "5ms"
`)
	eng := &richEngine{fakeEngine: fakeEngine{works: map[string]reconciler.WorkStatus{}}}
	mgr, err := managed.NewReconcilerManager(home, "", eng, managed.DefaultStore)
	if err != nil {
		t.Fatal(err)
	}
	if err := mgr.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	waitFor(t, func() bool { return eng.StartCount() > 0 })
	eng.mu.Lock()
	eng.failLoad = true
	eng.mu.Unlock()
	time.Sleep(80 * time.Millisecond)
	stopManager(t, mgr)
}
