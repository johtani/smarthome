package action

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel"
	"time"
)

// NoOpAction represents an action that does nothing for a specified interval.
type NoOpAction struct {
	interval time.Duration
}

// Run executes the NoOpAction by sleeping for the specified interval.
func (a NoOpAction) Run(ctx context.Context, _ string) (string, error) {
	_, span := otel.Tracer("action").Start(ctx, "NoOpAction.Run")
	defer span.End()
	time.Sleep(a.interval)
	return fmt.Sprintf("Paused for %v", a.interval), nil
}

// NewNoOpAction creates a new NoOpAction with the specified interval.
func NewNoOpAction(interval time.Duration) NoOpAction {
	return NoOpAction{
		interval: interval,
	}
}
