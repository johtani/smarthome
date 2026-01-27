package action

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel"
	"time"
)

type NoOpAction struct {
	interval time.Duration
}

func (a NoOpAction) Run(ctx context.Context, _ string) (string, error) {
	_, span := otel.Tracer("action").Start(ctx, "NoOpAction.Run")
	defer span.End()
	time.Sleep(a.interval)
	return fmt.Sprintf("Paused for %v", a.interval), nil
}

func NewNoOpAction(interval time.Duration) NoOpAction {
	return NoOpAction{
		interval: interval,
	}
}
