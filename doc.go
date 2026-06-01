// Package agentd provides domain helpers for user-scoped agent daemons: home layout
// defaults, run environment keys, harness registry, and managed-agent reconciliation.
//
// Generic local JSON-RPC daemon plumbing lives in github.com/brandonkramer/locdaemon.
// Import locdaemon subpackages directly for client dial/call, runtime loop, and observe
// topic names:
//
//	import (
//	    locclient "github.com/brandonkramer/locdaemon/client"
//	    loclayout "github.com/brandonkramer/locdaemon/layout"
//	    locruntime "github.com/brandonkramer/locdaemon/runtime"
//	    locobserve "github.com/brandonkramer/locdaemon/observe"
//	    "github.com/brandonkramer/agentd"
//	)
//
// Subpackages: harness (named run implementations), managed (catalog, state, reconciler).
// Root exports daemon home layout (Default, Agentd) and run env helpers (ChildEnv, VerifyRunID).
//
// Shared libraries: github.com/brandonkramer/locdaemon, github.com/brandonkramer/catalogfile,
// github.com/brandonkramer/ipc, github.com/brandonkramer/message, github.com/brandonkramer/poll,
// github.com/brandonkramer/reconciler, github.com/brandonkramer/jsonfile,
// github.com/brandonkramer/filelock, github.com/brandonkramer/registry,
// github.com/brandonkramer/svcroot, github.com/brandonkramer/procenv,
// github.com/brandonkramer/shellquote.
package agentd
