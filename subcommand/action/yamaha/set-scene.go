package yamaha

import "fmt"

type SetSceneAction struct {
	name  string
	scene int
	c     *Client
}

func (a SetSceneAction) Run() (string, error) {
	err := a.c.SetScene(a.scene)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Set scene to No.%v.", a.scene), nil
}

func NewSetSceneAction(client *Client) SetSceneAction {
	return SetSceneAction{
		"Set Yamaha Scene",
		70,
		client,
	}
}
