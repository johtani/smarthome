package switchbot

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel"
)

type ExecuteSceneAction struct {
	name    string
	sceneId string
	client  *CachedClient
}

func (a ExecuteSceneAction) Run(ctx context.Context, _ string) (string, error) {
	ctx, span := otel.Tracer("switchbot").Start(ctx, "ExecuteSceneAction.Run")
	defer span.End()
	err := a.client.SceneAPI.Execute(ctx, a.sceneId)
	if err != nil {
		return "", err
	}
	name, err := a.client.GetSceneName(ctx, a.sceneId)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Execute the scene(%v).", name), nil
}

func NewExecuteSceneAction(client *CachedClient, sceneId string) ExecuteSceneAction {
	return ExecuteSceneAction{
		name:    "Execute the scene",
		sceneId: sceneId,
		client:  client,
	}
}
