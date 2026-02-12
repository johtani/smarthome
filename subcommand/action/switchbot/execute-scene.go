package switchbot

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel"
)

// ExecuteSceneAction represents an action to execute a SwitchBot scene.
type ExecuteSceneAction struct {
	name    string
	sceneID string
	client  *CachedClient
}

// Run executes the ExecuteSceneAction.
func (a ExecuteSceneAction) Run(ctx context.Context, _ string) (string, error) {
	ctx, span := otel.Tracer("switchbot").Start(ctx, "ExecuteSceneAction.Run")
	defer span.End()
	err := a.client.SceneAPI.Execute(ctx, a.sceneID)
	if err != nil {
		return "", err
	}
	name, err := a.client.GetSceneName(ctx, a.sceneID)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Execute the scene(%v).", name), nil
}

// NewExecuteSceneAction creates a new ExecuteSceneAction.
func NewExecuteSceneAction(client *CachedClient, sceneId string) ExecuteSceneAction {
	return ExecuteSceneAction{
		name:    "Execute the scene",
		sceneID: sceneId,
		client:  client,
	}
}
