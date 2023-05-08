package switchbot

import (
	"context"
	"fmt"
	"github.com/nasa9084/go-switchbot/v2"
)

type SendCommandAction struct {
	name     string
	deviceId string
	command  switchbot.Command
	*switchbot.Client
}

func (a SendCommandAction) Run() (string, error) {
	err := a.Device().Command(context.Background(), a.deviceId, a.command)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Sned the command(%v) to the device(%v)", a.command, a.deviceId), nil
}

func NewSendCommandAction(client *switchbot.Client, deviceId string, command switchbot.Command) SendCommandAction {
	return SendCommandAction{
		"Send the command to the device on SwitchBot",
		deviceId,
		command,
		client,
	}
}
