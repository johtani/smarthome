package action

import (
	"fmt"
	"time"
)

type NoOpAction struct {
	interval time.Duration
}

func (a NoOpAction) Run(_ string) (string, error) {
	time.Sleep(a.interval)
	return fmt.Sprintf("Paused for %v", a.interval), nil
}

func NewNoOpAction(interval time.Duration) NoOpAction {
	return NoOpAction{
		interval: interval,
	}
}
