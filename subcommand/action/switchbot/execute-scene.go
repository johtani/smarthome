package switchbot

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel"
)

type ExecuteSceneAction struct {
	name    string
	sceneId string
	CachedClient
}

func (a ExecuteSceneAction) Run(ctx context.Context, _ string) (string, error) {
	ctx, span := otel.Tracer("switchbot").Start(ctx, "ExecuteSceneAction.Run")
	defer span.End()
	err := a.Scene().Execute(ctx, a.sceneId)
	if err != nil {
		return "", err
	}
	name, err := a.GetSceneName(ctx, a.sceneId)
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
