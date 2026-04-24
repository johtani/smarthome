package owntone

import (
	"context"
	"io"
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
		name                string
		query               string
		aliases             map[string]string
		expressionStatus    int
		expressionResponse  string
		queryStatus         int
		wantExpressionParts []string
		wantQueryText       string
	}{
		{
			name:               "expression search uses original and normalized keywords",
			query:              "ＭＧＡ!!!",
			aliases:            map[string]string{"mga": "Mrs GREEN APPLE"},
			expressionStatus:   http.StatusOK,
			expressionResponse: searchSampleJSONResponse(),
			queryStatus:        http.StatusOK,
			wantExpressionParts: []string{
				"title includes \"ＭＧＡ!!!\"",
				"title includes \"mrs green apple\"",
			},
			wantQueryText: "",
		},
		{
			name:               "fallback to query when expression search returns error",
			query:              "ＭＧＡ!!! type:track",
			aliases:            map[string]string{"mga": "Mrs GREEN APPLE"},
			expressionStatus:   http.StatusBadRequest,
			expressionResponse: `{"message":"invalid expression"}`,
			queryStatus:        http.StatusOK,
			wantExpressionParts: []string{
				"title includes \"ＭＧＡ!!!\"",
				"title includes \"mrs green apple\"",
			},
			wantQueryText: "mrs green apple",
		},
		{
			name:               "fallback to query when expression search has no hit",
			query:              "!!! type:artist",
			aliases:            nil,
			expressionStatus:   http.StatusOK,
			expressionResponse: emptySearchJSONResponse(),
			queryStatus:        http.StatusOK,
			wantExpressionParts: []string{
				"artist includes \"!!!\"",
			},
			wantQueryText: "!!!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var receivedExpressions []string
			var receivedQueries []string
			mux := http.NewServeMux()
			mux.HandleFunc("/api/search", func(w http.ResponseWriter, r *http.Request) {
				receivedExpression := r.URL.Query().Get("expression")
				receivedQuery := r.URL.Query().Get("query")
				receivedExpressions = append(receivedExpressions, receivedExpression)
				receivedQueries = append(receivedQueries, receivedQuery)
				if receivedExpression != "" {
					w.WriteHeader(tt.expressionStatus)
					_, _ = w.Write([]byte(tt.expressionResponse))
					return
				}
				w.WriteHeader(tt.queryStatus)
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

			expression := ""
			for _, v := range receivedExpressions {
				if v != "" {
					expression = v
					break
				}
			}
			for _, part := range tt.wantExpressionParts {
				if !strings.Contains(expression, part) {
					t.Fatalf("expression = %q, want to contain %q", expression, part)
				}
			}

			query := ""
			for _, v := range receivedQueries {
				if v != "" {
					query = v
				}
			}
			if query != tt.wantQueryText {
				t.Fatalf("search query = %q, want %q", query, tt.wantQueryText)
			}
		})
	}
}

func TestBuildSearchExpression(t *testing.T) {
	tests := []struct {
		name     string
		keywords []string
		types    []SearchType
		want     string
	}{
		{
			name:     "track type expression",
			keywords: []string{"Utada", "宇多田ひかる"},
			types:    []SearchType{track},
			want:     "(title includes \"Utada\" or artist includes \"Utada\" or album includes \"Utada\") or (title includes \"宇多田ひかる\" or artist includes \"宇多田ひかる\" or album includes \"宇多田ひかる\")",
		},
		{
			name:     "escape quote and slash",
			keywords: []string{`A"B\C`},
			types:    []SearchType{artist},
			want:     `(artist includes "A\"B\\C")`,
		},
		{
			name:     "empty keywords",
			keywords: nil,
			types:    []SearchType{artist},
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildSearchExpression(tt.keywords, tt.types)
			if got != tt.want {
				t.Fatalf("buildSearchExpression() = %q, want %q", got, tt.want)
			}
		})
	}
}

func emptySearchJSONResponse() string {
	return `{
  "tracks": {"items": [], "total": 0, "offset": 0, "limit": 5},
  "artists": {"items": [], "total": 0, "offset": 0, "limit": 5},
  "albums": {"items": [], "total": 0, "offset": 0, "limit": 5},
  "genres": {"items": [], "total": 0, "offset": 0, "limit": 5},
  "playlists": {"items": [], "total": 0, "offset": 0, "limit": 5}
}`
}
