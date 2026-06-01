package managed

import (
	"path/filepath"

	"github.com/brandonkramer/catalogfile"
)

// Store resolves managed-agent storage paths under a daemon home and project root.
type Store = catalogfile.Paths

// DefaultProjectDir is the conventional project-local config directory name.
const DefaultProjectDir = ".agentd"

// DefaultStore is the conventional managed-agent storage layout.
var DefaultStore = Store{
	GlobalCatalog:  func(home string) string { return filepath.Join(home, "agents.toml") },
	ProjectCatalog: func(projectRoot string) string { return filepath.Join(projectRoot, DefaultProjectDir, "agents.toml") },
	StateDir:       func(home string) string { return filepath.Join(home, "managed") },
	SecretsDir:     func(home string) string { return filepath.Join(home, "secrets") },
}
