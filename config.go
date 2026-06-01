package agentd

import "github.com/brandonkramer/svcroot"

type (
	// Layout names files under a daemon home directory.
	Layout = svcroot.Layout
	// Service bundles home resolution and layout defaults.
	Service = svcroot.Service
)

// Default is the conventional agentd layout: home/sessions/{daemon.sock,daemon.lock}.
var Default = Layout{
	SessionsDir:       "sessions",
	SocketName:        "daemon.sock",
	ObserveSocketName: "observe.sock",
	LockName:          "daemon.lock",
	PipePrefix:        "agentd",
}

// Agentd configures home resolution for the agentd daemon.
var Agentd = Service{
	EnvVar:     "AGENTD_HOME",
	DefaultDir: ".agentd",
	Layout:     Default,
}

// WorkLayout is the on-disk layout for agentd-work homes (~/.agentd-work).
var WorkLayout = func() Layout {
	l := Default
	l.PipePrefix = "agentd-work"
	return l
}()

// Work configures home resolution for the agentd-work CLI/daemon.
var Work = Service{
	EnvVar:       "AGENTD_HOME",
	DefaultDir:   ".agentd-work",
	Layout:       WorkLayout,
	RegistryFile: "sessions/known-daemons.json",
}
