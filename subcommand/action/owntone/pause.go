package owntone

import (
	"context"
	"github.com/johtani/smarthome/subcommand/action"
)

// PauseAction represents an action to pause playback on Owntone.
type PauseAction struct {
	name string
	c    *Client
}

// Run executes the PauseAction.
func (a PauseAction) Run(ctx context.Context, args string) (string, error) {
	ctx, span := action.StartRunSpan(ctx, "owntone", "PauseAction.Run", args)
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
