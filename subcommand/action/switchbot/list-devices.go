package switchbot

import (
	"context"
	"fmt"
	"strings"
)

type ListDevicesAction struct {
	name string
	CachedClient
}

func (a ListDevicesAction) Run(ctx context.Context, _ string) (string, error) {
	var msg []string
	pdev, vdev, err := a.Device().List(ctx)
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

func NewListDevicesAction(client CachedClient) ListDevicesAction {
	return ListDevicesAction{
		name:         "List devices on SwitchBot",
		CachedClient: client,
	}
}
