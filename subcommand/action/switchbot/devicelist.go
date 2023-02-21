package switchbot

import (
	"context"
	"fmt"
	"github.com/nasa9084/go-switchbot/v2"
)

type DeviceListAction struct {
	name string
	c    *switchbot.Client
}

func (a DeviceListAction) Run() error {
	pdev, vdev, err := a.c.Device().List(context.Background())
	if err != nil {
		return err
	}
	for _, d := range pdev {
		fmt.Printf("%s\t%s\n", d.Type, d.Name)
	}
	for _, d := range vdev {
		fmt.Printf("%s\t%s\n", d.Type, d.Name)
	}
	return nil
}

func NewDeviceListAction() DeviceListAction {
	return DeviceListAction{
		"List devices on SwitchBot",
		NewSwitchBotClient(),
	}
}
