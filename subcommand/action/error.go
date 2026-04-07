package action

import (
	"context"

	"go.opentelemetry.io/otel"
)

// ErrorAction always returns the configured error when executed.
type ErrorAction struct {
	err error
}

// Run executes the ErrorAction and returns the configured error.
func (a ErrorAction) Run(ctx context.Context, _ string) (string, error) {
	_, span := otel.Tracer("action").Start(ctx, "ErrorAction.Run")
	defer span.End()
	return "", a.err
}

// NewErrorAction creates a new ErrorAction.
func NewErrorAction(err error) ErrorAction {
	return ErrorAction{err: err}
}
