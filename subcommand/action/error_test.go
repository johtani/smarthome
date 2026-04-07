package action

import (
	"errors"
	"testing"
)

func TestErrorAction_Run(t *testing.T) {
	wantErr := errors.New("boom")
	a := NewErrorAction(wantErr)

	msg, err := a.Run(t.Context(), "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected wrapped error %v, got %v", wantErr, err)
	}
	if msg != "" {
		t.Fatalf("expected empty message, got %q", msg)
	}
}
