package owntone

import (
	"context"
	"go.opentelemetry.io/otel"
)

type PauseAction struct {
	name string
	c    *Client
}

func (a PauseAction) Run(ctx context.Context, _ string) (string, error) {
	ctx, span := otel.Tracer("owntone").Start(ctx, "PauseAction.Run")
	defer span.End()
	err := a.c.Pause(ctx)
	if err != nil {
		return "", err
	}
	return "Paused the music.", nil
}

func NewPauseAction(client *Client) PauseAction {
	return PauseAction{
		name: "Pause music on Owntone",
		c:    client,
	}
}
