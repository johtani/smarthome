package action

import (
	"context"
	"go.opentelemetry.io/otel"
)

type HelpAction struct {
	help string
}

func (a HelpAction) Run(ctx context.Context, _ string) (string, error) {
	_, span := otel.Tracer("action").Start(ctx, "HelpAction.Run")
	defer span.End()
	return a.help, nil
}

func NewHelpAction(msg string) HelpAction {
	return HelpAction{help: msg}
}
