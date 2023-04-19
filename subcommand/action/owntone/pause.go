package owntone

import "fmt"

type PauseAction struct {
	name string
	c    *Client
}

func (a PauseAction) Run() error {
	err := a.c.Pause()
	if err != nil {
		return err
	}
	fmt.Println("owntone pause action succeeded.")
	return nil
}

func NewPauseAction(client *Client) PauseAction {
	return PauseAction{
		"Pause music on Owntone",
		client,
	}
}
