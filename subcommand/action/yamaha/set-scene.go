package yamaha

import (
	"context"
	"fmt"
)

type SetSceneAction struct {
	name  string
	scene int
	c     *Client
}

func (a SetSceneAction) Run(ctx context.Context, _ string) (string, error) {
	err := a.c.SetScene(ctx, a.scene)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Set scene to No.%v.", a.scene), nil
}

func NewSetSceneAction(client *Client, scene int) SetSceneAction {
	return SetSceneAction{
		name:  "Set Yamaha Scene",
		scene: scene,
		c:     client,
	}
}
