package harness

import (
	"fmt"

	"github.com/brandonkramer/shellquote"
)

type genericCommand struct{}

func (genericCommand) Name() string { return GenericCommand }

func (genericCommand) Prepare(in *WorkInput) (Prepared, error) {
	if in.CommandTemplate == "" {
		return Prepared{}, fmt.Errorf("generic-command harness: %w", errCommandRequired)
	}
	path, err := shellquote.WritePrompt(in.RunDir, in.PromptContent)
	if err != nil {
		return Prepared{}, fmt.Errorf("generic-command harness: write prompt: %w", err)
	}
	return Prepared{
		Driver: GenericCommand, Harness: in.Harness,
		CommandTemplate: in.CommandTemplate,
		Command:         shellquote.SubstitutePrompt(in.CommandTemplate, path),
		PromptPath:      path,
	}, nil
}

// PrepareGeneric prepares a generic-command work unit.
func PrepareGeneric(in *WorkInput) (Prepared, error) {
	return genericCommand{}.Prepare(in)
}
