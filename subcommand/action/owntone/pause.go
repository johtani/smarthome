package owntone

import (
	"context"
	"go.opentelemetry.io/otel"
)

// PauseAction represents an action to pause playback on Owntone.
type PauseAction struct {
	name string
	c    *Client
}

// Run executes the PauseAction.
func (a PauseAction) Run(ctx context.Context, _ string) (string, error) {
	ctx, span := otel.Tracer("owntone").Start(ctx, "PauseAction.Run")
	defer span.End()
	err := a.c.Pause(ctx)
	if err != nil {
		return "", err
	}
	return "Paused the music.", nil
}

// NewPauseAction creates a new PauseAction.
func NewPauseAction(client *Client) PauseAction {
	return PauseAction{
		name: "Pause music on Owntone",
		c:    client,
	}
}
