package switchbot

import (
	"context"
	"fmt"
	"strings"

	"go.opentelemetry.io/otel"
)

type ListScenesAction struct {
	name string
	CachedClient
}

func (a ListScenesAction) Run(ctx context.Context, _ string) (string, error) {
	ctx, span := otel.Tracer("switchbot").Start(ctx, "ListScenesAction.Run")
	defer span.End()
	scenes, err := a.SceneAPI.List(ctx)
	var msg []string
	if err != nil {
		return "", err
	}
	for _, s := range scenes {
		msg = append(msg, fmt.Sprintf("%s\t%s", s.Name, s.ID))
	}
	return strings.Join(msg, "\n"), nil
}

func NewListScenesAction(client CachedClient) ListScenesAction {
	return ListScenesAction{
		name:         "List scenes on SwitchBot",
		CachedClient: client,
	}
}
