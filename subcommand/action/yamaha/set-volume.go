package yamaha

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel"
)

type SetVolumeAction struct {
	name   string
	volume int
	c      YamahaAPI
}

func (a SetVolumeAction) Run(ctx context.Context, _ string) (string, error) {
	ctx, span := otel.Tracer("yamaha").Start(ctx, "SetVolumeAction.Run")
	defer span.End()
	err := a.c.SetVolume(ctx, a.volume)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Set volume to %v.", a.volume), nil
}

func NewSetVolumeAction(client YamahaAPI) SetVolumeAction {
	return SetVolumeAction{
		name:   "Set Yamaha Volume",
		volume: 70,
		c:      client,
	}
}
