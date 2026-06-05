package harness

import "errors"

// ResolveDriver picks the effective driver from an explicit request, agent profile
// driver, or generic-command when a command template is present.
func ResolveDriver(requested, agentDriver, commandTemplate string) (string, error) {
	if requested != "" {
		return requested, nil
	}
	if agentDriver != "" {
		return agentDriver, nil
	}
	if commandTemplate != "" {
		return GenericCommand, nil
	}
	return "", errors.New("driver is required")
}
