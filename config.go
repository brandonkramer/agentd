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
