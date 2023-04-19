package switchbot

import (
	"context"
	"github.com/nasa9084/go-switchbot/v2"
)

type SendCommandAction struct {
	name     string
	deviceId string
	command  switchbot.Command
	*switchbot.Client
}

func (a SendCommandAction) Run() error {
	err := a.Device().Command(context.Background(), a.deviceId, a.command)
	if err != nil {
		return err
	}
	return nil
}

func NewSendCommandAction(client *switchbot.Client, deviceId string, command switchbot.Command) SendCommandAction {
	return SendCommandAction{
		"List scenes on SwitchBot",
		deviceId,
		command,
		client,
	}
}
