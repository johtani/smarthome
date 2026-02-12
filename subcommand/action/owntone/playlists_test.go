package owntone

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDisplayPlaylistsAction_Run(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/library/playlists", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"items": [
				{"uri": "library:playlist:1", "name": "My Playlist", "item_count": 1, "path": "library:playlist:1"},
				{"uri": "spotify:playlist:2", "name": "Spotify Playlist", "item_count": 1, "path": "spotify:playlist:2"},
				{"uri": "library:playlist:3", "name": "Another One", "item_count": 1, "path": "library:playlist:3"}
			]
		}`))
	})

	server := httptest.NewServer(mux)
	defer server.Close()
	client := NewClient(Config{URL: server.URL})
	action := NewDisplayPlaylistsAction(client)

	got, err := action.Run(context.Background(), "")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if !strings.Contains(got, "My Playlist") || !strings.Contains(got, "Another One") || !strings.Contains(got, "Spotify Playlist (by Spotify)") {
		t.Errorf("Run() result missing expected playlists: %s", got)
	}
}

func TestDisplayPlaylistsAction_Run_OnlySpotify(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/library/playlists", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"items": [
				{"uri": "library:playlist:1", "name": "My Playlist", "item_count": 1, "path": "library:playlist:1"},
				{"uri": "spotify:playlist:2", "name": "Spotify Playlist", "item_count": 1, "path": "spotify:playlist:2"}
			]
		}`))
	})

	server := httptest.NewServer(mux)
	defer server.Close()
	client := NewClient(Config{URL: server.URL})
	action := NewDisplayPlaylistsAction(client)

	got, err := action.Run(context.Background(), "spotify")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if !strings.Contains(got, "Spotify Playlist (by Spotify)") {
		t.Errorf("Run() result missing spotify playlist: %s", got)
	}
	if strings.Contains(got, "  My Playlist") {
		t.Errorf("Run() result should not contain library playlist when category is spotify: %s", got)
	}
}
