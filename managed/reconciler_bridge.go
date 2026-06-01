package managed

import (
	"context"
	"time"

	"github.com/brandonkramer/catalogfile"
	"github.com/brandonkramer/reconciler"
)

// Engine reconciles managed agents for the generic reconciler.
type Engine interface {
	StartWork(ctx context.Context, spec *AgentSpec, state *State) (workID string, err error)
	ResolveBinding(spec *AgentSpec, state *State) (*TaskBinding, error)
	LoadWork(workID string) (reconciler.WorkStatus, error)
	GracefulStopWork(workID string, grace time.Duration)
	ShuttingDown() bool
	NowISO() string
}

type agentWorkEngine struct {
	eng Engine
}

func (e *agentWorkEngine) StartWork(ctx context.Context, spec *AgentSpec, state *reconciler.State) (string, error) {
	var ms *State
	if state != nil && state.Name != "" {
		s := fromReconcilerState(*state)
		ms = &s
	}
	return e.eng.StartWork(ctx, spec, ms)
}

func (e *agentWorkEngine) LoadWork(workID string) (reconciler.WorkStatus, error) {
	return e.eng.LoadWork(workID)
}

func (e *agentWorkEngine) GracefulStopWork(workID string, grace time.Duration) {
	e.eng.GracefulStopWork(workID, grace)
}

func (e *agentWorkEngine) ShuttingDown() bool { return e.eng.ShuttingDown() }

func (e *agentWorkEngine) NowISO() string { return e.eng.NowISO() }

type agentStore struct {
	home  string
	store Store
}

func (s *agentStore) Load(name string) (reconciler.State, error) {
	st, err := LoadState(s.home, name, s.store)
	if err != nil {
		return reconciler.State{}, err
	}
	return toReconcilerState(st), nil
}

func (s *agentStore) Save(st *reconciler.State) error {
	ms := fromReconcilerState(*st)
	return SaveState(s.home, &ms, s.store)
}

func (s *agentStore) Delete(name string) error {
	return DeleteState(s.home, name, s.store)
}

// ErrInvalidUnit is returned when a reconciler unit lacks managed agent data.
var ErrInvalidUnit = agentErr("invalid managed unit")

type agentErr string

func (e agentErr) Error() string { return "agentd/managed: " + string(e) }

// NewReconcilerManager wires managed catalog/store into the generic reconciler.
func NewReconcilerManager(home, projectRoot string, engine Engine, store Store) (*reconciler.Manager, error) {
	return catalogfile.NewFromFiles(catalogfile.Config[AgentSpec]{
		Home:        home,
		ProjectRoot: projectRoot,
		Paths:       store,
		ParseFile:   loadFile,
		ToUnit:      unitFromSpec,
		Store:       &agentStore{home: home, store: store},
		WorkEngine:  &agentWorkEngine{eng: engine},
		Interval:    DefaultReconcilerInterval,
		Hash: func(home string, spec *AgentSpec, st *reconciler.State) (string, error) {
			var ms *State
			if st != nil && st.Name != "" {
				s := fromReconcilerState(*st)
				ms = &s
			}
			bind, err := engine.ResolveBinding(spec, ms)
			if err != nil {
				return "", err
			}
			return HashSpec(home, spec, bind, store)
		},
	})
}

func unitFromSpec(name string, spec AgentSpec) reconciler.Unit {
	specPtr := spec
	specPtr.Name = name
	return reconciler.Unit{
		Name:         name,
		Enabled:      spec.Enabled,
		Restart:      reconciler.ParseRestartPolicy(spec.Restart),
		MaxRestarts:  spec.MaxRestarts,
		Backoff:      spec.Backoff,
		GracePeriod:  spec.GracePeriod,
		PollInterval: spec.PollInterval,
		Data:         &specPtr,
	}
}

func toReconcilerState(st State) reconciler.State {
	return reconciler.State{
		Name:          st.Name,
		DesiredHash:   st.DesiredHash,
		AppliedHash:   st.AppliedHash,
		RunID:         st.RunID,
		Restarts:      st.Restarts,
		LastExitCode:  st.LastExitCode,
		LastError:     st.LastError,
		LastStartedAt: st.LastStartedAt,
		Converging:    st.Converging,
	}
}

func fromReconcilerState(st reconciler.State) State {
	return State{
		Name:          st.Name,
		DesiredHash:   st.DesiredHash,
		AppliedHash:   st.AppliedHash,
		RunID:         st.RunID,
		Restarts:      st.Restarts,
		LastExitCode:  st.LastExitCode,
		LastError:     st.LastError,
		LastStartedAt: st.LastStartedAt,
		Converging:    st.Converging,
	}
}
