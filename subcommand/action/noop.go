package action

import "time"

type NoOpAction struct {
	interval time.Duration
}

func (a NoOpAction) Run() error {
	time.Sleep(a.interval)
	return nil
}

func NewNoOpAction(interval time.Duration) NoOpAction {
	return NoOpAction{
		interval,
	}
}
