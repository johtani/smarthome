package yamaha

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel"
)

type SetInputAction struct {
	name  string
	input string
	c     YamahaAPI
}

func (a SetInputAction) Run(ctx context.Context, _ string) (string, error) {
	ctx, span := otel.Tracer("yamaha").Start(ctx, "SetInputAction.Run")
	defer span.End()
	err := a.c.SetInput(ctx, a.input)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Set input to %v.", a.input), nil
}

func NewSetInputAction(client YamahaAPI, input string) SetInputAction {
	return SetInputAction{
		name:  "Set Yamaha Input",
		input: input,
		c:     client,
	}
}
