package yamaha

import (
	"context"
	"go.opentelemetry.io/otel"
)

// PowerOffAction represents an action to turn off the Yamaha device.
type PowerOffAction struct {
	name string
	c    API
}

// Run executes the PowerOffAction.
func (a PowerOffAction) Run(ctx context.Context, _ string) (string, error) {
	ctx, span := otel.Tracer("yamaha").Start(ctx, "PowerOffAction.Run")
	defer span.End()
	err := a.c.PowerOff(ctx)
	if err != nil {
		return "", err
	}
	return "Amplifier Power off.", nil
}

// NewPowerOffAction creates a new PowerOffAction.
func NewPowerOffAction(client API) PowerOffAction {
	return PowerOffAction{
		name: "Power off Yamaha Amplifier",
		c:    client,
	}
}
