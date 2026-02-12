package yamaha

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel"
)

// SetSceneAction represents an action to set the scene on the Yamaha device.
type SetSceneAction struct {
	name  string
	scene int
	c     API
}

// Run executes the SetSceneAction.
func (a SetSceneAction) Run(ctx context.Context, _ string) (string, error) {
	ctx, span := otel.Tracer("yamaha").Start(ctx, "SetSceneAction.Run")
	defer span.End()
	err := a.c.SetScene(ctx, a.scene)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Set scene to No.%v.", a.scene), nil
}

// NewSetSceneAction creates a new SetSceneAction.
func NewSetSceneAction(client API, scene int) SetSceneAction {
	return SetSceneAction{
		name:  "Set Yamaha Scene",
		scene: scene,
		c:     client,
	}
}
