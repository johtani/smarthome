package owntone

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHTTPMusicIntentResolver_Resolve_DirectIntent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"artist_candidates":["宇多田ヒカル"],"confidence":0.9}`))
	}))
	defer server.Close()

	r := NewHTTPMusicIntentResolver(server.URL, 2*time.Second)
	intent, err := r.Resolve(context.Background(), "宇多田ヒカルを再生")
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if len(intent.ArtistCandidates) != 1 || intent.ArtistCandidates[0] != "宇多田ヒカル" {
		t.Fatalf("unexpected intent: %+v", intent)
	}
}

func TestHTTPMusicIntentResolver_Resolve_WrappedIntent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"music_intent":{"track_candidates":["First Love"],"confidence":0.8},"model":"gpt-test","reason":"wrapped"}`))
	}))
	defer server.Close()

	r := NewHTTPMusicIntentResolver(server.URL, 2*time.Second)
	intent, err := r.Resolve(context.Background(), "First Love")
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if len(intent.TrackCandidates) != 1 || intent.TrackCandidates[0] != "First Love" {
		t.Fatalf("unexpected intent: %+v", intent)
	}
	if intent.Model != "gpt-test" {
		t.Fatalf("model = %q, want %q", intent.Model, "gpt-test")
	}
	if intent.Reason != "wrapped" {
		t.Fatalf("reason = %q, want %q", intent.Reason, "wrapped")
	}
}

func TestHTTPMusicIntentResolver_Resolve_Non200(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "unavailable", http.StatusServiceUnavailable)
	}))
	defer server.Close()

	r := NewHTTPMusicIntentResolver(server.URL, 2*time.Second)
	_, err := r.Resolve(context.Background(), "test")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
