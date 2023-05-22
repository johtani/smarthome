package yamaha

import "fmt"

type SetVolumeAction struct {
	name   string
	volume int
	c      *Client
}

func (a SetVolumeAction) Run() (string, error) {
	err := a.c.SetVolume(a.volume)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Set volume to %v.", a.volume), nil
}

func NewSetVolumeAction(client *Client) SetVolumeAction {
	return SetVolumeAction{
		"Set Yamaha Volume",
		70,
		client,
	}
}
