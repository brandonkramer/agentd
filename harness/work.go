package harness

const (
	// GenericCommand is the shell-template driver name.
	GenericCommand = "generic-command"
	// ClaudeCode is the Claude CLI driver name.
	ClaudeCode = "claude-code"
)

// WorkInput is everything a harness needs to prepare a run.
type WorkInput struct {
	WorkDir         string
	RunDir          string
	CommandTemplate string
	PromptContent   string
	Model           string
	Harness         string
}

// Prepared is the resolved run command and harness metadata.
type Prepared struct {
	Driver          string
	Harness         string
	CommandTemplate string
	Command         string
	PromptPath      string
	ExecPath        string
	ExecArgs        []string
	ExecDir         string
	StdinPrompt     bool
	Warnings        []string
}
