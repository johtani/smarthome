package action

import (
	"context"
	"go.opentelemetry.io/otel"
)

// HelpAction represents an action that displays the help message.
type HelpAction struct {
	help string
}

// Run executes the HelpAction by returning the help string.
func (a HelpAction) Run(ctx context.Context, _ string) (string, error) {
	_, span := otel.Tracer("action").Start(ctx, "HelpAction.Run")
	defer span.End()
	return a.help, nil
}

// NewHelpAction creates a new HelpAction with the given help message.
func NewHelpAction(msg string) HelpAction {
	return HelpAction{help: msg}
}
