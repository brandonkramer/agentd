# agentd

Helpers for local-scoped agent daemons.

## Install

```bash
go get github.com/brandonkramer/agentd
```

## Usage

Resolve the daemon home (from `AGENTD_HOME` or `~/.agentd`):

```go
home, err := agentd.Agentd.Resolve()
if err != nil {
    return err
}
```

Mark child processes with run metadata:

```go
env := agentd.ChildEnv(home, runID, map[string]string{
    "WORK_DIR": workDir,
})
```

Prepare a headless run (built-in harnesses: `generic-command`, `claude-code`):

```go
h, err := harness.Get("generic-command")
if err != nil {
    return err
}

prep, err := h.Prepare(&harness.WorkInput{
    RunDir:          runDir,
    WorkDir:         workDir,
    CommandTemplate: "claude -p < {prompt}",
    PromptContent:   prompt,
})
if err != nil {
    return err
}
// prep.Command, prep.ExecPath, prep.ExecArgs, prep.PromptPath
```

Load managed agents from `~/agents.toml` and project `.agentd/agents.toml`:

```go
managed.SetDriverValidator(func(name string) error {
    _, err := harness.Get(name)
    return err
})

specs, err := managed.LoadMergedSpecs(home, projectRoot, managed.DefaultStore)
if err != nil {
    return err
}

mgr, err := managed.NewReconcilerManager(home, projectRoot, engine, managed.DefaultStore)
if err != nil {
    return err
}
```

## Development

```bash
make check
```
