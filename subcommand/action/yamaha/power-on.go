package yamaha

import (
	"context"

	"go.opentelemetry.io/otel"
)

// PowerOnAction represents an action to turn on the Yamaha device.
type PowerOnAction struct {
	name string
	c    API
}

// Run executes the PowerOnAction.
func (a PowerOnAction) Run(ctx context.Context, _ string) (string, error) {
	ctx, span := otel.Tracer("yamaha").Start(ctx, "PowerOnAction.Run")
	defer span.End()
	err := a.c.PowerOn(ctx)
	if err != nil {
		return "", err
	}
	return "Power on Yamaha.", nil
}

// NewPowerOnAction creates a new PowerOnAction.
func NewPowerOnAction(client API) PowerOnAction {
	return PowerOnAction{
		name: "Power On Yamaha",
		c:    client,
	}
}
