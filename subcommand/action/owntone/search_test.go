package owntone

import (
	"context"
	"fmt"
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

func TestNormalizeTextNoKana(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "NFKC and lowercase", in: "ＴＥＳＴ　１２３", want: "test 123"},
		{name: "Symbol collapse", in: "A・B!!! C", want: "a b c"},
		{name: "Katakana kept as-is", in: "スピッツ", want: "スピッツ"},
		{name: "Mixed kana not converted", in: "ヒカル ひかる", want: "ヒカル ひかる"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeTextNoKana(tt.in)
			if got != tt.want {
				t.Fatalf("normalizeTextNoKana() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNormalizeSearchKeywordForExternalSearch(t *testing.T) {
	tests := []struct {
		name    string
		keyword string
		aliases map[string]string
		want    string
	}{
		{
			name:    "katakana kept without kana conversion",
			keyword: "スピッツ",
			aliases: nil,
			want:    "スピッツ",
		},
		{
			name:    "alias match without kana conversion",
			keyword: "MGA",
			aliases: map[string]string{"ｍｇａ": "Mrs. GREEN APPLE"},
			want:    "mrs green apple",
		},
		{
			name:    "alias key in katakana matched without conversion",
			keyword: "ヒッキー",
			aliases: map[string]string{"ヒッキー": "宇多田ヒカル"},
			want:    "宇多田ヒカル",
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
			got := normalizeSearchKeywordForExternalSearch(tt.keyword, tt.aliases)
			if got != tt.want {
				t.Fatalf("normalizeSearchKeywordForExternalSearch() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSearchAndPlayAction_Run_ExternalSearchUsesNoKanaKeyword(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/queue/clear", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("/api/queue/items/add", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	external := &fakeExternalSearcher{
		result: &SearchResult{
			Tracks: Items{
				Items: []SearchItem{{Title: "Robinson", Artist: "スピッツ", URI: "library:track:1"}},
				Total: 1,
			},
		},
	}
	client := NewClient(Config{URL: server.URL})
	action := NewSearchAndPlayAction(client, WithExternalSearch(external))

	_, err := action.Run(context.Background(), "スピッツ")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if external.receivedKeyword != "スピッツ" {
		t.Fatalf("external search keyword = %q, want %q (katakana must not be converted to hiragana)", external.receivedKeyword, "スピッツ")
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

func TestBuildSearchExpressionAND(t *testing.T) {
	got := buildSearchExpressionAND([]string{"Utada", "First Love"}, []SearchType{track})
	want := "(title includes \"Utada\" or artist includes \"Utada\" or album includes \"Utada\") and (title includes \"First Love\" or artist includes \"First Love\" or album includes \"First Love\")"
	if got != want {
		t.Fatalf("buildSearchExpressionAND() = %q, want %q", got, want)
	}
}

type fakeMusicIntentResolver struct {
	intent MusicIntent
	err    error
}

func (f fakeMusicIntentResolver) Path() string {
	return "fake_music_intent"
}

func (f fakeMusicIntentResolver) Resolve(_ context.Context, _ string) (MusicIntent, error) {
	if f.err != nil {
		return MusicIntent{}, f.err
	}
	return f.intent, nil
}

type fakeExternalSearcher struct {
	result          *SearchResult
	err             error
	calls           int
	receivedKeyword string
}

func (f *fakeExternalSearcher) Search(_ context.Context, keyword string, _ []SearchType, _ int) (*SearchResult, error) {
	f.calls++
	f.receivedKeyword = keyword
	if f.err != nil {
		return nil, f.err
	}
	return f.result, nil
}

func TestSearchAndPlayAction_Run_ExternalSearchSuccessUsesURIPlayback(t *testing.T) {
	var searchCalled bool
	var addURIs string
	mux := http.NewServeMux()
	mux.HandleFunc("/api/search", func(w http.ResponseWriter, _ *http.Request) {
		searchCalled = true
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, searchSampleJSONResponse())
	})
	mux.HandleFunc("/api/queue/clear", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("/api/queue/items/add", func(w http.ResponseWriter, r *http.Request) {
		addURIs = r.URL.Query().Get("uris")
		w.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	external := &fakeExternalSearcher{
		result: &SearchResult{
			Tracks: Items{
				Items: []SearchItem{{Title: "First Love", Artist: "宇多田ヒカル", URI: "library:track:123"}},
				Total: 1,
			},
		},
	}
	client := NewClient(Config{URL: server.URL})
	action := NewSearchAndPlayAction(client, WithExternalSearch(external))

	got, err := action.Run(context.Background(), "宇多田 type:track")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !strings.Contains(got, "And play these items") {
		t.Fatalf("Run() result = %q, want play message", got)
	}
	if addURIs != "library:track:123" {
		t.Fatalf("uris = %q, want library:track:123", addURIs)
	}
	if searchCalled {
		t.Fatal("expected Owntone search not to be called when external search succeeds")
	}
	if external.calls != 1 {
		t.Fatalf("external calls = %d, want 1", external.calls)
	}
}

func TestSearchAndPlayAction_Run_ExternalSearchFailureFallsBackToOwntone(t *testing.T) {
	var expressionSearchCalled bool
	var addQueueCalled bool
	mux := http.NewServeMux()
	mux.HandleFunc("/api/search", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("expression") != "" {
			expressionSearchCalled = true
		}
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, searchSampleJSONResponse())
	})
	mux.HandleFunc("/api/queue/clear", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("/api/queue/items/add", func(w http.ResponseWriter, _ *http.Request) {
		addQueueCalled = true
		w.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	external := &fakeExternalSearcher{err: fmt.Errorf("external down")}
	client := NewClient(Config{URL: server.URL})
	action := NewSearchAndPlayAction(client, WithExternalSearch(external))

	got, err := action.Run(context.Background(), "宇多田 type:track")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !strings.Contains(got, "And play these items") {
		t.Fatalf("Run() result = %q, want play message", got)
	}
	if !expressionSearchCalled {
		t.Fatal("expected Owntone expression search fallback")
	}
	if !addQueueCalled {
		t.Fatal("expected queue add after fallback search")
	}
}

func TestSearchAndPlayAction_Run_MusicIntentStrictSkipsExternalSearch(t *testing.T) {
	external := &fakeExternalSearcher{
		result: &SearchResult{
			Tracks: Items{Items: []SearchItem{{Title: "External", URI: "library:track:external"}}, Total: 1},
		},
	}
	var addURIs string
	mux := http.NewServeMux()
	mux.HandleFunc("/api/search", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("expression") == "" {
			t.Fatal("expected strict expression search")
		}
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, searchSampleJSONResponse())
	})
	mux.HandleFunc("/api/queue/clear", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("/api/queue/items/add", func(w http.ResponseWriter, r *http.Request) {
		addURIs = r.URL.Query().Get("uris")
		w.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(mux)
	defer server.Close()
	client := NewClient(Config{URL: server.URL})
	action := NewSearchAndPlayAction(
		client,
		WithExternalSearch(external),
		WithMusicIntentResolver(fakeMusicIntentResolver{
			intent: MusicIntent{TrackCandidates: []string{"First Love"}, Confidence: 0.95},
		}),
	)

	_, err := action.Run(context.Background(), "First Loveかけて")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if external.calls != 0 {
		t.Fatalf("external calls = %d, want 0", external.calls)
	}
	if addURIs == "library:track:external" {
		t.Fatal("strict intent should not play external search result")
	}
}

func TestSearchAndPlayAction_Run_LowConfidenceShowsCandidatesWithoutPlaying(t *testing.T) {
	var receivedExpression string
	var clearQueueCalled bool
	var addQueueCalled bool

	mux := http.NewServeMux()
	mux.HandleFunc("/api/search", func(w http.ResponseWriter, r *http.Request) {
		receivedExpression = r.URL.Query().Get("expression")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, searchSampleJSONResponse())
	})
	mux.HandleFunc("/api/queue/clear", func(w http.ResponseWriter, _ *http.Request) {
		clearQueueCalled = true
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("/api/queue/items/add", func(w http.ResponseWriter, _ *http.Request) {
		addQueueCalled = true
		w.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(mux)
	defer server.Close()
	client := NewClient(Config{URL: server.URL})
	action := NewSearchAndPlayAction(
		client,
		WithMusicIntentResolver(fakeMusicIntentResolver{
			intent: MusicIntent{
				ArtistCandidates: []string{"宇多田ヒカル"},
				TrackCandidates:  []string{"First Love"},
				Confidence:       0.40,
			},
		}),
		WithMusicIntentConfidenceThreshold(0.75),
	)

	got, err := action.Run(context.Background(), "宇多田ヒカルのFirst Loveかけて")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !strings.Contains(got, "Candidate results only (no autoplay).") {
		t.Fatalf("Run() result = %q, want candidate-only message", got)
	}
	if !strings.Contains(receivedExpression, " and ") {
		t.Fatalf("strict expression = %q, want AND expression", receivedExpression)
	}
	if clearQueueCalled || addQueueCalled {
		t.Fatalf("expected no playback actions, clear=%v add=%v", clearQueueCalled, addQueueCalled)
	}
}

func TestSearchAndPlayAction_Run_StrictGenreUsesIntentGenreForPlayback(t *testing.T) {
	var addExpression string
	mux := http.NewServeMux()
	mux.HandleFunc("/api/search", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("expression") != "" {
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, `{
  "tracks": {"items": [], "total": 0, "offset": 0, "limit": 5},
  "artists": {"items": [], "total": 0, "offset": 0, "limit": 5},
  "albums": {"items": [], "total": 0, "offset": 0, "limit": 5},
  "genres": {"items": [{"name":"Rock","track_count":10}], "total": 1, "offset": 0, "limit": 5},
  "playlists": {"items": [], "total": 0, "offset": 0, "limit": 5}
}`)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, emptySearchJSONResponse())
	})
	mux.HandleFunc("/api/queue/clear", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("/api/queue/items/add", func(w http.ResponseWriter, r *http.Request) {
		addExpression = r.URL.Query().Get("expression")
		w.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(mux)
	defer server.Close()
	client := NewClient(Config{
		URL:           server.URL,
		SearchAliases: map[string]string{"ろっく": "rock"},
	})
	action := NewSearchAndPlayAction(
		client,
		WithMusicIntentResolver(fakeMusicIntentResolver{
			intent: MusicIntent{
				GenreCandidates: []string{"ロック"},
				Confidence:      0.95,
			},
		}),
	)

	_, err := action.Run(context.Background(), "ロック系をかけて")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if addExpression != `genre is "rock"` {
		t.Fatalf("expression = %q, want %q", addExpression, `genre is "rock"`)
	}
}

func TestSearchAndPlayAction_Run_AmbiguousIntentShowsCandidatesWithoutPlaying(t *testing.T) {
	var addQueueCalled bool
	mux := http.NewServeMux()
	mux.HandleFunc("/api/search", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, searchSampleJSONResponse())
	})
	mux.HandleFunc("/api/queue/clear", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("/api/queue/items/add", func(w http.ResponseWriter, _ *http.Request) {
		addQueueCalled = true
		w.WriteHeader(http.StatusOK)
	})

	server := httptest.NewServer(mux)
	defer server.Close()
	client := NewClient(Config{URL: server.URL})
	action := NewSearchAndPlayAction(
		client,
		WithMusicIntentResolver(fakeMusicIntentResolver{
			intent: MusicIntent{
				TrackCandidates: []string{"First Love"},
				Confidence:      0.99,
				Ambiguous:       true,
			},
		}),
	)

	got, err := action.Run(context.Background(), "First Loveかけて")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !strings.Contains(got, "Candidate results only (no autoplay).") {
		t.Fatalf("Run() result = %q, want candidate-only message", got)
	}
	if addQueueCalled {
		t.Fatal("expected no playback actions for ambiguous intent")
	}
}

func TestWithMusicIntentConfidenceThreshold_AllowsZero(t *testing.T) {
	action := NewSearchAndPlayAction(
		NewClient(Config{URL: "http://example.local"}),
		WithMusicIntentConfidenceThreshold(0),
	)
	if action.musicIntentConfidenceThreshold != 0 {
		t.Fatalf("threshold = %v, want 0", action.musicIntentConfidenceThreshold)
	}
}

func TestSearchAndPlayAction_Run_MusicIntentFailureFallsBackToLegacy(t *testing.T) {
	var searchCalls int
	mux := http.NewServeMux()
	mux.HandleFunc("/api/search", func(w http.ResponseWriter, r *http.Request) {
		searchCalls++
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, searchSampleJSONResponse())
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
	action := NewSearchAndPlayAction(
		client,
		WithMusicIntentResolver(fakeMusicIntentResolver{err: fmt.Errorf("boom")}),
	)

	got, err := action.Run(context.Background(), "keyword")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if !strings.Contains(got, "And play these items") {
		t.Fatalf("unexpected result: %q", got)
	}
	if searchCalls == 0 {
		t.Fatal("expected legacy search to be called")
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
