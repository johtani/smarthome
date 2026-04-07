package yamaha

import (
	"context"
	"github.com/johtani/smarthome/subcommand/action"
)

// PowerOffAction represents an action to turn off the Yamaha device.
type PowerOffAction struct {
	name string
	c    API
}

// Run executes the PowerOffAction.
func (a PowerOffAction) Run(ctx context.Context, args string) (string, error) {
	ctx, span := action.StartRunSpan(ctx, "yamaha", "PowerOffAction.Run", args)
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
