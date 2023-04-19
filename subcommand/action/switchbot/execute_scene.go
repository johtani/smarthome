package switchbot

import (
	"context"
	"github.com/nasa9084/go-switchbot/v2"
)

type ExecuteSceneAction struct {
	name    string
	sceneId string
	*switchbot.Client
}

func (a ExecuteSceneAction) Run() error {
	err := a.Scene().Execute(context.Background(), a.sceneId)
	if err != nil {
		return err
	}
	return nil
}

func NewExecuteSceneAction(client *switchbot.Client, sceneId string) ExecuteSceneAction {
	return ExecuteSceneAction{
		"Turn off the target device",
		sceneId,
		client,
	}
}
