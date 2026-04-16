package resolver

import "testing"

func TestEnsureRequestID(t *testing.T) {
	ctx, requestID := EnsureRequestID(t.Context())
	if requestID == "" {
		t.Fatal("expected non-empty request ID")
	}

	got, ok := RequestIDFromContext(ctx)
	if !ok {
		t.Fatal("expected request ID in context")
	}
	if got != requestID {
		t.Fatalf("expected request ID %q, got %q", requestID, got)
	}
}

func TestChannelContextHelpers(t *testing.T) {
	ctx := WithChannel(t.Context(), "slack_mention")
	channel, ok := ChannelFromContext(ctx)
	if !ok {
		t.Fatal("expected channel in context")
	}
	if channel != "slack_mention" {
		t.Fatalf("expected slack_mention, got %q", channel)
	}
}
