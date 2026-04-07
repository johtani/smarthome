package action

import (
	"context"
)

// ErrorAction always returns the configured error when executed.
type ErrorAction struct {
	err error
}

// Run executes the ErrorAction and returns the configured error.
func (a ErrorAction) Run(ctx context.Context, args string) (string, error) {
	_, span := StartRunSpan(ctx, "action", "ErrorAction.Run", args)
	defer span.End()
	return "", a.err
}

// NewErrorAction creates a new ErrorAction.
func NewErrorAction(err error) ErrorAction {
	return ErrorAction{err: err}
}
