package switchbot

import (
	"context"
	"fmt"
	"strings"

	"go.opentelemetry.io/otel"
)

type ListDevicesAction struct {
	name   string
	client *CachedClient
}

func (a ListDevicesAction) Run(ctx context.Context, _ string) (string, error) {
	ctx, span := otel.Tracer("switchbot").Start(ctx, "ListDevicesAction.Run")
	defer span.End()
	var msg []string
	pdev, vdev, err := a.client.DeviceAPI.List(ctx)
	if err != nil {
		return "", err
	}
	for _, d := range pdev {
		msg = append(msg, fmt.Sprintf("%s\t%s\t%s", d.Type, d.Name, d.ID))
	}
	for _, d := range vdev {
		msg = append(msg, fmt.Sprintf("%s\t%s\t%s", d.Type, d.Name, d.ID))
	}
	return strings.Join(msg, "\n"), nil
}

func NewListDevicesAction(client *CachedClient) ListDevicesAction {
	return ListDevicesAction{
		name:   "List devices on SwitchBot",
		client: client,
	}
}
