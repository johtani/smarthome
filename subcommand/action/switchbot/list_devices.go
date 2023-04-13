package switchbot

import (
	"context"
	"fmt"
	"github.com/nasa9084/go-switchbot/v2"
)

type ListDevicesAction struct {
	name string
	c    *switchbot.Client
}

func (a ListDevicesAction) Run() error {
	pdev, vdev, err := a.c.Device().List(context.Background())
	if err != nil {
		return err
	}
	for _, d := range pdev {
		fmt.Printf("%s\t%s\t%s\n", d.Type, d.Name, d.ID)
	}
	for _, d := range vdev {
		fmt.Printf("%s\t%s\t%s\n", d.Type, d.Name, d.ID)
	}
	return nil
}

func NewListDevicesAction(client *switchbot.Client) ListDevicesAction {
	return ListDevicesAction{
		"List devices on SwitchBot",
		client,
	}
}
