package owntone

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	type args struct {
		target string
	}
	tests := []struct {
		name    string
		args    args
		want    *SearchQuery
		wantErr bool
	}{
		{name: "Term only", args: args{target: "term"}, want: &SearchQuery{Terms: []string{"term"}, Offset: -1, Limit: -1}},
		{name: "2 Terms", args: args{target: "日本語 twice"}, want: &SearchQuery{Terms: []string{"日本語", "twice"}, Offset: -1, Limit: -1}},
		{name: "Term and offset", args: args{target: "term offset:1"}, want: &SearchQuery{Terms: []string{"term"}, Offset: 1, Limit: -1}},
		{name: "Term and offset, limit", args: args{target: "term offset:1 limit:2"}, want: &SearchQuery{Terms: []string{"term"}, Offset: 1, Limit: 2}},
		{name: "Term and offset, limit, types", args: args{target: "term offset:1 limit:2 type:album type:artist"}, want: &SearchQuery{Terms: []string{"term"}, Offset: 1, Limit: 2, Types: []SearchType{album, artist}}},
		{name: "Term genre type", args: args{target: "term type:genre"}, want: &SearchQuery{Terms: []string{"term"}, Offset: -1, Limit: -1, Types: []SearchType{genre}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Parse(tt.args.target); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSearchAndDisplayAction_Run(t *testing.T) {
	server := createMockServerWithResponse(http.StatusOK, http.MethodGet, "/api/search", nil, searchSampleJSONResponse())
	defer server.Close()
	client := NewClient(Config{URL: server.URL})
	action := NewSearchAndDisplayAction(client)

	got, err := action.Run(context.Background(), "keyword")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	expectedParts := []string{
		"Search Results...",
		"# Artists (1)",
		" The xx",
		"# Albums (3)",
		" Away From the Sun / 3 Doors Down",
		" The Better Life / 3 Doors Down",
		"# Tracks (14)",
		" Another Love / Tom Odell",
		" Away From the Sun / 3 Doors Down",
		"# Genres (182)",
		" Abstract",
		" Alternative",
	}

	for _, part := range expectedParts {
		if !strings.Contains(got, part) {
			t.Errorf("Run() result does not contain %q\nGot:\n%s", part, got)
		}
	}
}

func TestSearchAndPlayAction_Run(t *testing.T) {
	// We need multiple endpoints to be mocked.
	// But our createMockServer is limited to one path.
	// We can use a custom mux.
	mux := http.NewServeMux()
	mux.HandleFunc("/api/search", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(searchSampleJSONResponse()))
	})
	mux.HandleFunc("/api/queue/clear", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("/api/queue/items/add", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(mux)
	defer server.Close()
	client := NewClient(Config{URL: server.URL})
	action := NewSearchAndPlayAction(client)

	got, err := action.Run(context.Background(), "keyword")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if !strings.Contains(got, "And play these items") {
		t.Errorf("Run() result does not contain expected success message, got: %s", got)
	}
}
