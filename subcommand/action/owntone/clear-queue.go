package owntone

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel"
)

// ClearQueueAction represents an action to clear the playback queue on Owntone.
type ClearQueueAction struct {
	name string
	c    *Client
}

// Run executes the ClearQueueAction.
func (a ClearQueueAction) Run(ctx context.Context, _ string) (string, error) {
	ctx, span := otel.Tracer("owntone").Start(ctx, "ClearQueueAction.Run")
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
