package yamaha

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel"
)

// SetVolumeAction represents an action to set the volume on the Yamaha device.
type SetVolumeAction struct {
	name   string
	volume int
	c      YamahaAPI
}

// Run executes the SetVolumeAction.
func (a SetVolumeAction) Run(ctx context.Context, _ string) (string, error) {
	ctx, span := otel.Tracer("yamaha").Start(ctx, "SetVolumeAction.Run")
	defer span.End()
	err := a.c.SetVolume(ctx, a.volume)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Set volume to %v.", a.volume), nil
}

// NewSetVolumeAction creates a new SetVolumeAction.
func NewSetVolumeAction(client YamahaAPI, volume int) SetVolumeAction {
	return SetVolumeAction{
		name:   "Set Yamaha Volume",
		volume: volume,
		c:      client,
	}
}
