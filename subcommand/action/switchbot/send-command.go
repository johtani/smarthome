package switchbot

import (
	"context"
	"fmt"
	"github.com/nasa9084/go-switchbot/v3"
	"go.opentelemetry.io/otel"
)

type SendCommandAction struct {
	name     string
	deviceID string
	command  switchbot.Command
	client   *CachedClient
}

func (a SendCommandAction) Run(ctx context.Context, _ string) (string, error) {
	ctx, span := otel.Tracer("switchbot").Start(ctx, "SendCommandAction.Run")
	defer span.End()
	err := a.client.DeviceAPI.Command(ctx, a.deviceID, a.command)
	if err != nil {
		return "", err
	}
	name, err := a.client.GetDeviceName(ctx, a.deviceID)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Send the command(%v) to the device(%v)", a.command, name), nil
}

func NewSendCommandAction(client *CachedClient, deviceId string, command switchbot.Command) SendCommandAction {
	return SendCommandAction{
		name:     "Send the command to the device on SwitchBot",
		deviceID: deviceId,
		command:  command,
		client:   client,
	}
}
