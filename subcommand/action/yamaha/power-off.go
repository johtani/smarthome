package yamaha

import (
	"context"
	"go.opentelemetry.io/otel"
)

type PowerOffAction struct {
	name string
	c    YamahaAPI
}

func (a PowerOffAction) Run(ctx context.Context, _ string) (string, error) {
	ctx, span := otel.Tracer("yamaha").Start(ctx, "PowerOffAction.Run")
	defer span.End()
	err := a.c.PowerOff(ctx)
	if err != nil {
		return "", err
	}
	return "Amplifier Power off.", nil
}

func NewPowerOffAction(client YamahaAPI) PowerOffAction {
	return PowerOffAction{
		name: "Power off Yamaha Amplifier",
		c:    client,
	}
}
