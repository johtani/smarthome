package switchbot

import (
	"context"
	"fmt"
	"github.com/nasa9084/go-switchbot/v2"
)

type ExecuteSceneAction struct {
	name    string
	sceneId string
	*switchbot.Client
}

func (a ExecuteSceneAction) Run() (string, error) {
	err := a.Scene().Execute(context.Background(), a.sceneId)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Execute the scene(%v).", a.sceneId), nil
}

func NewExecuteSceneAction(client *switchbot.Client, sceneId string) ExecuteSceneAction {
	return ExecuteSceneAction{
		"Execute the scene",
		sceneId,
		client,
	}
}
