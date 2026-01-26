package action

import (
	"context"
	"testing"
	"time"
)

func TestNoOpAction_Run(t *testing.T) {
	interval := 10 * time.Millisecond
	a := NewNoOpAction(interval)

	start := time.Now()
	got, err := a.Run(context.Background(), "")
	elapsed := time.Since(start)

	if err != nil {
		t.Errorf("Run() error = %v, wantErr %v", err, false)
		return
	}

	want := "Paused for 10ms"
	if got != want {
		t.Errorf("Run() got = %v, want %v", got, want)
	}

	if elapsed < interval {
		t.Errorf("Run() took %v, want at least %v", elapsed, interval)
	}
}
