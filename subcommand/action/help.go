package action

import (
	"context"
)

// HelpAction represents an action that displays the help message.
type HelpAction struct {
	help string
}

// Run executes the HelpAction by returning the help string.
func (a HelpAction) Run(ctx context.Context, args string) (string, error) {
	_, span := StartRunSpan(ctx, "action", "HelpAction.Run", args)
	defer span.End()
	return a.help, nil
}

// NewHelpAction creates a new HelpAction with the given help message.
func NewHelpAction(msg string) HelpAction {
	return HelpAction{help: msg}
}
