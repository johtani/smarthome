package action

import "context"

// FixedArgsAction wraps another action and always passes fixed args to it.
type FixedArgsAction struct {
	action Action
	args   string
}

// Run executes the wrapped action with fixed args.
func (a FixedArgsAction) Run(ctx context.Context, args string) (string, error) {
	return a.action.Run(ctx, a.args)
}

// NewFixedArgsAction creates a new FixedArgsAction.
func NewFixedArgsAction(action Action, args string) FixedArgsAction {
	return FixedArgsAction{
		action: action,
		args:   args,
	}
}
