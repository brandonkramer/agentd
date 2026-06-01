package agentd

import (
	locrunenv "github.com/brandonkramer/locdaemon/runenv"
	"github.com/brandonkramer/proctree"
)

const (
	// EnvRunID marks a child process with its run id.
	EnvRunID = "AGENTD_RUN_ID"
	// EnvHome marks a child process with its daemon home.
	EnvHome = "AGENTD_HOME"
)

var agentEnvKeys = locrunenv.EnvKeys{RunID: EnvRunID, Home: EnvHome}

// ChildEnv returns os.Environ plus AGENTD_RUN_ID, AGENTD_HOME, and extra keys.
func ChildEnv(home, runID string, extra map[string]string) []string {
	return locrunenv.ChildEnv(agentEnvKeys, home, runID, extra)
}

// VerifyRunID reports whether pid refers to runID, using env markers or fallback spec.
func VerifyRunID(runID string, pid int, alive func(int) bool, fallback *proctree.Spec) bool {
	return locrunenv.VerifyRunID(agentEnvKeys, runID, pid, alive, fallback)
}
