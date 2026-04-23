package owntone

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
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
		{name: "Term genre type misspelled", args: args{target: "term type:gener"}, want: &SearchQuery{Terms: []string{"term"}, Offset: -1, Limit: -1, Types: []SearchType{genre}}},
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

func TestNormalizeText(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "NFKC and lowercase", in: "ＴＥＳＴ　１２３", want: "test 123"},
		{name: "Symbol collapse", in: "A・B!!! C", want: "a b c"},
		{name: "Kana normalize", in: "ヒカル ひかる", want: "ひかる ひかる"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeText(tt.in)
			if got != tt.want {
				t.Fatalf("normalizeText() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNormalizeSearchKeyword(t *testing.T) {
	tests := []struct {
		name    string
		keyword string
		aliases map[string]string
		want    string
	}{
		{
			name:    "alias with normalization",
			keyword: "MGA",
			aliases: map[string]string{"ｍｇａ": "Mrs. GREEN APPLE"},
			want:    "mrs green apple",
		},
		{
			name:    "term alias",
			keyword: "宇多田 ヒッキー",
			aliases: map[string]string{"ヒッキー": "宇多田ヒカル"},
			want:    "宇多田 宇多田ひかる",
		},
		{
			name:    "empty after normalization",
			keyword: "!!!",
			aliases: map[string]string{"x": "y"},
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeSearchKeyword(tt.keyword, tt.aliases)
			if got != tt.want {
				t.Fatalf("normalizeSearchKeyword() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSearchAndPlayAction_Run_UsesNormalizedQueryAndFallback(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		aliases        map[string]string
		wantSearchText string
	}{
		{
			name:           "normalized alias query",
			query:          "ＭＧＡ!!!",
			aliases:        map[string]string{"mga": "Mrs GREEN APPLE"},
			wantSearchText: "mrs green apple",
		},
		{
			name:           "fallback to original terms when normalized empty",
			query:          "!!! type:artist",
			aliases:        nil,
			wantSearchText: "!!!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var receivedQuery string
			mux := http.NewServeMux()
			mux.HandleFunc("/api/search", func(w http.ResponseWriter, r *http.Request) {
				receivedQuery = r.URL.Query().Get("query")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(searchSampleJSONResponse()))
			})
			mux.HandleFunc("/api/queue/clear", func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			})
			mux.HandleFunc("/api/queue/items/add", func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = io.WriteString(w, "")
			})

			server := httptest.NewServer(mux)
			defer server.Close()
			client := NewClient(Config{URL: server.URL, SearchAliases: tt.aliases})
			action := NewSearchAndPlayAction(client)

			_, err := action.Run(context.Background(), tt.query)
			if err != nil {
				t.Fatalf("Run() error = %v", err)
			}

			got, err := url.QueryUnescape(receivedQuery)
			if err != nil {
				t.Fatalf("QueryUnescape() error = %v", err)
			}
			if got != tt.wantSearchText {
				t.Fatalf("search query = %q, want %q", got, tt.wantSearchText)
			}
		})
	}
}
