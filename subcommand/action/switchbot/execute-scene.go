package switchbot

import (
	"context"
	"fmt"
)

type ExecuteSceneAction struct {
	name    string
	sceneId string
	CachedClient
}

func (a ExecuteSceneAction) Run(_ string) (string, error) {
	err := a.Scene().Execute(context.Background(), a.sceneId)
	if err != nil {
		return "", err
	}
	name, err := a.GetSceneName(a.sceneId)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Execute the scene(%v).", name), nil
}

func NewExecuteSceneAction(client CachedClient, sceneId string) ExecuteSceneAction {
	return ExecuteSceneAction{
		name:         "Execute the scene",
		sceneId:      sceneId,
		CachedClient: client,
	}
}
