package action

import (
	"fmt"
	"time"
)

import "context"

type NoOpAction struct {
	interval time.Duration
}

func (a NoOpAction) Run(_ context.Context, _ string) (string, error) {
	time.Sleep(a.interval)
	return fmt.Sprintf("Paused for %v", a.interval), nil
}

func NewNoOpAction(interval time.Duration) NoOpAction {
	return NoOpAction{
		interval: interval,
	}
}
