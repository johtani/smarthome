package yamaha

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel"
)

// SetInputAction represents an action to set the input source on the Yamaha device.
type SetInputAction struct {
	name  string
	input string
	c     API
}

// Run executes the SetInputAction.
func (a SetInputAction) Run(ctx context.Context, _ string) (string, error) {
	ctx, span := otel.Tracer("yamaha").Start(ctx, "SetInputAction.Run")
	defer span.End()
	err := a.c.SetInput(ctx, a.input)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Set input to %v.", a.input), nil
}

// NewSetInputAction creates a new SetInputAction.
func NewSetInputAction(client API, input string) SetInputAction {
	return SetInputAction{
		name:  "Set Yamaha Input",
		input: input,
		c:     client,
	}
}
