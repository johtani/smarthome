package owntone

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestOpenSearchExternalSearcher_Search(t *testing.T) {
	var received openSearchTemplateRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want %s", r.Method, http.MethodPost)
		}
		if r.URL.Path != "/smarthome-owntone-music/_search/template" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, openSearchSampleResponse())
	}))
	defer server.Close()

	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(trace.WithSyncer(exporter))
	previous := otel.GetTracerProvider()
	otel.SetTracerProvider(tp)
	defer func() {
		otel.SetTracerProvider(previous)
		_ = tp.Shutdown(context.Background())
	}()

	searcher := NewOpenSearchExternalSearcher(ExternalSearchConfig{
		OpenSearchURL: server.URL,
		Index:         "smarthome-owntone-music",
		TemplateID:    "music_search",
	})

	got, err := searcher.Search(context.Background(), "宇多田", []SearchType{track}, 2)
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}

	if received.ID != "music_search" {
		t.Fatalf("template id = %q", received.ID)
	}
	if received.Params.Query != "宇多田" {
		t.Fatalf("query = %q", received.Params.Query)
	}
	if !reflect.DeepEqual(received.Params.Types, []string{"track"}) {
		t.Fatalf("types = %v", received.Params.Types)
	}
	if !received.Params.HasTypes {
		t.Fatal("has_types = false, want true")
	}
	if received.Params.Size != 2 {
		t.Fatalf("size = %d", received.Params.Size)
	}

	want := &SearchResult{
		Tracks:  Items{Items: []SearchItem{{Title: "First Love", URI: "library:track:123", Artist: "宇多田ヒカル"}}, Total: 1, Limit: 1},
		Artists: Items{Items: []SearchItem{{URI: "library:artist:10", Name: "宇多田ヒカル", Artist: "宇多田ヒカル"}}, Total: 1, Limit: 1},
	}
	if !reflect.DeepEqual(got.Tracks, want.Tracks) {
		t.Fatalf("tracks = %+v, want %+v", got.Tracks, want.Tracks)
	}
	if !reflect.DeepEqual(got.Artists, want.Artists) {
		t.Fatalf("artists = %+v, want %+v", got.Artists, want.Artists)
	}

	spans := exporter.GetSpans()
	found := false
	for _, span := range spans {
		if span.Name == "owntone.ExternalSearch.Search" {
			found = true
			if attrValue(span.Attributes, "search.hit_count") != int64(2) {
				t.Fatalf("search.hit_count = %v, want 2", attrValue(span.Attributes, "search.hit_count"))
			}
			reqBody, ok := attrValue(span.Attributes, "search.request_body").(string)
			if !ok || reqBody == "" {
				t.Fatalf("search.request_body = %v, want non-empty string", attrValue(span.Attributes, "search.request_body"))
			}
			if !strings.Contains(reqBody, "宇多田") {
				t.Fatalf("search.request_body does not contain keyword: %s", reqBody)
			}
			if !strings.Contains(reqBody, "music_search") {
				t.Fatalf("search.request_body does not contain template id: %s", reqBody)
			}
		}
	}
	if !found {
		t.Fatal("expected owntone.ExternalSearch.Search span")
	}
}

func TestOpenSearchExternalSearcher_SearchSendsTypesArrayForDefaultTypes(t *testing.T) {
	var receivedRaw map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&receivedRaw); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, openSearchSampleResponse())
	}))
	defer server.Close()

	searcher := NewOpenSearchExternalSearcher(ExternalSearchConfig{
		OpenSearchURL: server.URL,
		Index:         "smarthome-owntone-music",
		TemplateID:    "music_search",
	})

	_, err := searcher.Search(context.Background(), "メイヤ", []SearchType{artist, album, track, genre}, 5)
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}

	params, ok := receivedRaw["params"].(map[string]any)
	if !ok {
		t.Fatalf("params = %#v", receivedRaw["params"])
	}
	if _, ok := params["type"]; ok {
		t.Fatalf("params.type is present: %#v", params["type"])
	}
	types, ok := params["types"].([]any)
	if !ok {
		t.Fatalf("params.types = %#v", params["types"])
	}
	gotTypes := make([]string, 0, len(types))
	for _, value := range types {
		gotTypes = append(gotTypes, value.(string))
	}
	wantTypes := []string{"artist", "album", "track", "genre"}
	if !reflect.DeepEqual(gotTypes, wantTypes) {
		t.Fatalf("params.types = %v, want %v", gotTypes, wantTypes)
	}
	if params["has_types"] != true {
		t.Fatalf("params.has_types = %#v, want true", params["has_types"])
	}
}

func TestOpenSearchExternalSearcher_SearchErrors(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		response   string
		wantReason string
	}{
		{name: "non 2xx", statusCode: http.StatusInternalServerError, response: `{}`, wantReason: "bad_status"},
		{name: "invalid json", statusCode: http.StatusOK, response: `{`, wantReason: "invalid_json"},
		{name: "no hits", statusCode: http.StatusOK, response: `{"hits":{"hits":[]}}`, wantReason: "no_hits"},
		{name: "missing uri", statusCode: http.StatusOK, response: `{"hits":{"hits":[{"_source":{"type":"track","title":"First Love"}}]}}`, wantReason: "no_hits"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.statusCode)
				_, _ = io.WriteString(w, tt.response)
			}))
			defer server.Close()

			searcher := NewOpenSearchExternalSearcher(ExternalSearchConfig{
				OpenSearchURL: server.URL,
				Index:         "smarthome-owntone-music",
				TemplateID:    "music_search",
			})
			_, err := searcher.Search(context.Background(), "宇多田", []SearchType{track}, 5)
			if err == nil {
				t.Fatal("expected error")
			}
			if reason := externalSearchFallbackReason(err); reason != tt.wantReason {
				t.Fatalf("fallback reason = %q, want %q (err=%v)", reason, tt.wantReason, err)
			}
		})
	}
}

func TestOpenSearchResponseToSearchResult_IgnoresUnsupportedAndMissingURI(t *testing.T) {
	response := openSearchTemplateResponse{}
	response.Hits.Hits = []openSearchHit{
		{Source: openSearchMusicDocument{URI: "", Type: "track", Title: "missing"}},
		{Source: openSearchMusicDocument{URI: "library:playlist:1", Type: "playlist", Title: "playlist"}},
		{Source: openSearchMusicDocument{URI: "library:album:1", Type: "album", Album: "First Love", Artist: "宇多田ヒカル"}},
	}

	got := openSearchResponseToSearchResult(response)
	if totalSearchResultCount(got) != 1 {
		t.Fatalf("total = %d, want 1", totalSearchResultCount(got))
	}
	if got.Albums.Items[0].Name != "First Love" {
		t.Fatalf("album name = %q", got.Albums.Items[0].Name)
	}
}

func attrValue(attrs []attribute.KeyValue, key string) any {
	for _, attr := range attrs {
		if string(attr.Key) == key {
			return attr.Value.AsInterface()
		}
	}
	return nil
}

func openSearchSampleResponse() string {
	return strings.TrimSpace(`{
  "hits": {
    "hits": [
      {
        "_source": {
          "uri": "library:track:123",
          "type": "track",
          "title": "First Love",
          "artist": "宇多田ヒカル",
          "album": "First Love",
          "genre": "J-Pop"
        }
      },
      {
        "_source": {
          "uri": "library:artist:10",
          "type": "artist",
          "title": "宇多田ヒカル",
          "artist": "宇多田ヒカル"
        }
      }
    ]
  }
}`)
}
