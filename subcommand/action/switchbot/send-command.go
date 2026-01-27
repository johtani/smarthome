package switchbot

import (
	"context"
	"fmt"
	"github.com/nasa9084/go-switchbot/v3"
	"go.opentelemetry.io/otel"
)

type SendCommandAction struct {
	name     string
	deviceId string
	command  switchbot.Command
	CachedClient
}

func (a SendCommandAction) Run(ctx context.Context, _ string) (string, error) {
	ctx, span := otel.Tracer("switchbot").Start(ctx, "SendCommandAction.Run")
	defer span.End()
	err := a.DeviceAPI.Command(ctx, a.deviceId, a.command)
	if err != nil {
		return "", err
	}
	name, err := a.GetDeviceName(ctx, a.deviceId)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Send the command(%v) to the device(%v)", a.command, name), nil
}

func NewSendCommandAction(client CachedClient, deviceId string, command switchbot.Command) SendCommandAction {
	return SendCommandAction{
		name:         "Send the command to the device on SwitchBot",
		deviceId:     deviceId,
		command:      command,
		CachedClient: client,
	}
}
