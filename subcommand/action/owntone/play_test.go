package owntone

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPlayAction_Run(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/player", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(playerStatusSampleJSONResponse()))
	})
	mux.HandleFunc("/api/library/playlists", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(playlistsSampleJSONResponse()))
	})
	mux.HandleFunc("/api/queue/items/add", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/api/player/play", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("/api/library", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(countsSampleJSONResponse()))
	})
	mux.HandleFunc("/api/library/artists", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(getArtistsSampleJSONResponse()))
	})
	mux.HandleFunc("/api/library/genres", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(getGenresSampleJSONResponse()))
	})

	server := httptest.NewServer(mux)
	defer server.Close()
	client := NewClient(Config{URL: server.URL})
	action := NewPlayAction(client)
	action.intn = func(n int) int {
		if n <= 1 {
			return 0
		}
		return n - 1
	}

	tests := []struct {
		name    string
		args    string
		want    string
		wantErr bool
	}{
		{
			name: "Play from playlist (queue is empty)",
			args: "",
			want: "Playing music from radio.",
		},
		{
			name: "Play random artist",
			args: "artist",
			want: "Add Artist : Ace Of Base",
		},
		{
			name: "Play random genre",
			args: "genre",
			want: "Add Genre : Alternative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := action.Run(context.Background(), tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Run() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPlayAction_Run_QueueNotEmpty(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/player", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"state": "pause", "item_id": 123}`))
	})
	mux.HandleFunc("/api/player/play", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	server := httptest.NewServer(mux)
	defer server.Close()
	client := NewClient(Config{URL: server.URL})
	action := NewPlayAction(client)

	got, err := action.Run(context.Background(), "")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if got != "Playing music  from queue" {
		t.Errorf("Run() got = %v, want %v", got, "Playing music  from queue")
	}
}
