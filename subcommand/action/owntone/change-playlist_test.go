package owntone

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestChangePlaylistAction_Run(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/library/playlists", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"items": [
				{"uri": "library:playlist:1", "name": "Morning", "item_count": 5, "path": "library:playlist:1"},
				{"uri": "library:playlist:2", "name": "Night", "item_count": 7, "path": "library:playlist:2"}
			]
		}`))
	})
	mux.HandleFunc("/api/queue/items/add", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(mux)
	defer server.Close()
	client := NewClient(Config{URL: server.URL})
	action := NewChangePlaylistAction(client)
	action.intn = func(_ int) int { return 1 }

	got, err := action.Run(context.Background(), "")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if got != "Change playlist to Night." {
		t.Errorf("Run() got = %v, want %v", got, "Change playlist to Night.")
	}
}
