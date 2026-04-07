package yamaha

import (
	"context"
	"fmt"
	"github.com/johtani/smarthome/subcommand/action"
)

// SetVolumeAction represents an action to set the volume on the Yamaha device.
type SetVolumeAction struct {
	name   string
	volume int
	c      API
}

// Run executes the SetVolumeAction.
func (a SetVolumeAction) Run(ctx context.Context, args string) (string, error) {
	ctx, span := action.StartRunSpan(ctx, "yamaha", "SetVolumeAction.Run", args)
	defer span.End()
	err := a.c.SetVolume(ctx, a.volume)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Set volume to %v.", a.volume), nil
}

// NewSetVolumeAction creates a new SetVolumeAction.
func NewSetVolumeAction(client API, volume int) SetVolumeAction {
	return SetVolumeAction{
		name:   "Set Yamaha Volume",
		volume: volume,
		c:      client,
	}
}
