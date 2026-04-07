package switchbot

import (
	"context"
	"fmt"
	"strings"

	"github.com/johtani/smarthome/subcommand/action"
)

// ListDevicesAction represents an action to list all SwitchBot devices.
type ListDevicesAction struct {
	name   string
	client *CachedClient
}

// Run executes the ListDevicesAction.
func (a ListDevicesAction) Run(ctx context.Context, args string) (string, error) {
	ctx, span := action.StartRunSpan(ctx, "switchbot", "ListDevicesAction.Run", args)
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

// NewListDevicesAction creates a new ListDevicesAction.
func NewListDevicesAction(client *CachedClient) ListDevicesAction {
	return ListDevicesAction{
		name:   "List devices on SwitchBot",
		client: client,
	}
}
