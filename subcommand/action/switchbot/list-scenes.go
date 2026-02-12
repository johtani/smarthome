package switchbot

import (
	"context"
	"fmt"
	"strings"

	"go.opentelemetry.io/otel"
)

// ListScenesAction represents an action to list all SwitchBot scenes.
type ListScenesAction struct {
	name   string
	client *CachedClient
}

// Run executes the ListScenesAction.
func (a ListScenesAction) Run(ctx context.Context, _ string) (string, error) {
	ctx, span := otel.Tracer("switchbot").Start(ctx, "ListScenesAction.Run")
	defer span.End()
	scenes, err := a.client.SceneAPI.List(ctx)
	var msg []string
	if err != nil {
		return "", err
	}
	for _, s := range scenes {
		msg = append(msg, fmt.Sprintf("%s\t%s", s.Name, s.ID))
	}
	return strings.Join(msg, "\n"), nil
}

// NewListScenesAction creates a new ListScenesAction.
func NewListScenesAction(client *CachedClient) ListScenesAction {
	return ListScenesAction{
		name:   "List scenes on SwitchBot",
		client: client,
	}
}
