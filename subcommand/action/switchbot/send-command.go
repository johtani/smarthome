package switchbot

import (
	"context"
	"fmt"
	"github.com/nasa9084/go-switchbot/v3"
)

type SendCommandAction struct {
	name     string
	deviceId string
	command  switchbot.Command
	CachedClient
}

func (a SendCommandAction) Run() (string, error) {
	err := a.Device().Command(context.Background(), a.deviceId, a.command)
	if err != nil {
		return "", err
	}
	name, err := a.GetDeviceName(a.deviceId)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Sned the command(%v) to the device(%v)", a.command, name), nil
}

func NewSendCommandAction(client CachedClient, deviceId string, command switchbot.Command) SendCommandAction {
	return SendCommandAction{
		name:         "Send the command to the device on SwitchBot",
		deviceId:     deviceId,
		command:      command,
		CachedClient: client,
	}
}
