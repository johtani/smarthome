/*
Package healthcheck provides actions for checking the health of various services.
*/
package healthcheck

import (
	"context"

	"github.com/johtani/smarthome/subcommand/action/owntone"
	"github.com/johtani/smarthome/subcommand/action/switchbot"
	"github.com/johtani/smarthome/subcommand/action/yamaha"
	"go.opentelemetry.io/otel"
)

// SwitchBotHealthCheckAction checks the health of the SwitchBot API.
type SwitchBotHealthCheckAction struct {
	client *switchbot.CachedClient
}

// Run executes the SwitchBotHealthCheckAction.
func (a SwitchBotHealthCheckAction) Run(ctx context.Context, _ string) (string, error) {
	ctx, span := otel.Tracer("healthcheck").Start(ctx, "SwitchBotHealthCheckAction.Run")
	defer span.End()
	_, _, err := a.client.DeviceAPI.List(ctx)
	if err != nil {
		return "SwitchBot: Error (" + err.Error() + ")", nil
	}
	return "SwitchBot: OK", nil
}

// NewSwitchBotHealthCheckAction creates a new SwitchBotHealthCheckAction.
func NewSwitchBotHealthCheckAction(client *switchbot.CachedClient) SwitchBotHealthCheckAction {
	return SwitchBotHealthCheckAction{client: client}
}

// OwnToneHealthCheckAction checks the health of the OwnTone API.
type OwnToneHealthCheckAction struct {
	client *owntone.Client
}

// Run executes the OwnToneHealthCheckAction.
func (a OwnToneHealthCheckAction) Run(ctx context.Context, _ string) (string, error) {
	ctx, span := otel.Tracer("healthcheck").Start(ctx, "OwnToneHealthCheckAction.Run")
	defer span.End()
	_, err := a.client.Counts(ctx)
	if err != nil {
		return "OwnTone: Error (" + err.Error() + ")", nil
	}
	return "OwnTone: OK", nil
}

// NewOwnToneHealthCheckAction creates a new OwnToneHealthCheckAction.
func NewOwnToneHealthCheckAction(client *owntone.Client) OwnToneHealthCheckAction {
	return OwnToneHealthCheckAction{client: client}
}

// YamahaHealthCheckAction checks the health of the Yamaha API.
type YamahaHealthCheckAction struct {
	client yamaha.API
}

// Run executes the YamahaHealthCheckAction.
func (a YamahaHealthCheckAction) Run(ctx context.Context, _ string) (string, error) {
	ctx, span := otel.Tracer("healthcheck").Start(ctx, "YamahaHealthCheckAction.Run")
	defer span.End()
	err := a.client.GetDeviceInfo(ctx)
	if err != nil {
		return "Yamaha: Error (" + err.Error() + ")", nil
	}
	return "Yamaha: OK", nil
}

// NewYamahaHealthCheckAction creates a new YamahaHealthCheckAction.
func NewYamahaHealthCheckAction(client yamaha.API) YamahaHealthCheckAction {
	return YamahaHealthCheckAction{client: client}
}
