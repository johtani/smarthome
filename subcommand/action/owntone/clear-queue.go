package owntone

import (
	"context"
	"fmt"
	"github.com/johtani/smarthome/subcommand/action"
)

// ClearQueueAction represents an action to clear the playback queue on Owntone.
type ClearQueueAction struct {
	name string
	c    *Client
}

// Run executes the ClearQueueAction.
func (a ClearQueueAction) Run(ctx context.Context, args string) (string, error) {
	ctx, span := action.StartRunSpan(ctx, "owntone", "ClearQueueAction.Run", args)
	defer span.End()
	err := a.c.ClearQueue(ctx)
	if err != nil {
		return "", fmt.Errorf("error in ClearQueue(%v)\n %v", a.c.config.URL, err)
	}
	return "Cleared queue", nil
}

// NewClearQueueAction creates a new ClearQueueAction.
func NewClearQueueAction(client *Client) ClearQueueAction {
	return ClearQueueAction{
		name: "Clear queue on Owntone",
		c:    client,
	}
}
