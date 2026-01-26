package owntone

import (
	"context"
	"fmt"
)

type SetVolumeAction struct {
	name   string
	volume int
	c      *Client
}

func (a SetVolumeAction) Run(ctx context.Context, _ string) (string, error) {
	err := a.c.SetVolume(ctx, a.volume)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Set volume to %v.", a.volume), nil
}

func NewSetVolumeAction(client *Client) SetVolumeAction {
	return SetVolumeAction{
		name:   "Set Owntone Volume",
		volume: 33,
		c:      client,
	}
}
