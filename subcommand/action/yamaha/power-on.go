package yamaha

import (
	"context"

	"go.opentelemetry.io/otel"
)

type PowerOnAction struct {
	name string
	c    YamahaAPI
}

func (a PowerOnAction) Run(ctx context.Context, _ string) (string, error) {
	ctx, span := otel.Tracer("yamaha").Start(ctx, "PowerOnAction.Run")
	defer span.End()
	err := a.c.PowerOn(ctx)
	if err != nil {
		return "", err
	}
	return "Power on Yamaha.", nil
}

func NewPowerOnAction(client YamahaAPI) PowerOnAction {
	return PowerOnAction{
		name: "Power On Yamaha",
		c:    client,
	}
}
