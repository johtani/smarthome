package yamaha

import "fmt"

type SetVolumeAction struct {
	name   string
	volume int
	c      *Client
}

func (a SetVolumeAction) Run(_ string) (string, error) {
	err := a.c.SetVolume(a.volume)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Set volume to %v.", a.volume), nil
}

func NewSetVolumeAction(client *Client) SetVolumeAction {
	return SetVolumeAction{
		name:   "Set Yamaha Volume",
		volume: 70,
		c:      client,
	}
}
