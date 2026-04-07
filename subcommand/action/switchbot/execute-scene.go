package switchbot

import (
	"context"
	"fmt"
	"github.com/johtani/smarthome/subcommand/action"
)

// ExecuteSceneAction represents an action to execute a SwitchBot scene.
type ExecuteSceneAction struct {
	name    string
	sceneID string
	client  *CachedClient
}

// Run executes the ExecuteSceneAction.
func (a ExecuteSceneAction) Run(ctx context.Context, args string) (string, error) {
	ctx, span := action.StartRunSpan(ctx, "switchbot", "ExecuteSceneAction.Run", args)
	defer span.End()
	err := a.client.Execute(ctx, a.sceneID)
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
