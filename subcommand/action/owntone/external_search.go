package owntone

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const defaultExternalSearchTimeout = 3 * time.Second

// ErrExternalSearchNoHits indicates that OpenSearch returned no playable hits.
var ErrExternalSearchNoHits = errors.New("external search returned no playable hits")

// ExternalSearcher searches an external music index and returns Owntone-shaped results.
type ExternalSearcher interface {
	Search(ctx context.Context, keyword string, resultTypes []SearchType, limit int) (*SearchResult, error)
}

// OpenSearchExternalSearcher calls OpenSearch Search Template API.
type OpenSearchExternalSearcher struct {
	config ExternalSearchConfig
	client http.Client
}

// NewOpenSearchExternalSearcher creates an OpenSearch-backed external search adapter.
func NewOpenSearchExternalSearcher(config ExternalSearchConfig) *OpenSearchExternalSearcher {
	timeout := defaultExternalSearchTimeout
	if config.TimeoutSeconds > 0 {
		timeout = time.Duration(config.TimeoutSeconds) * time.Second
	}
	return &OpenSearchExternalSearcher{
		config: config,
		client: http.Client{
			Timeout:   timeout,
			Transport: otelhttp.NewTransport(http.DefaultTransport),
		},
	}
}

type openSearchTemplateRequest struct {
	ID     string                   `json:"id"`
	Params openSearchTemplateParams `json:"params"`
}

type openSearchTemplateParams struct {
	Query string `json:"query"`
	Type  string `json:"type,omitempty"`
	Size  int    `json:"size"`
}

type openSearchTemplateResponse struct {
	Hits struct {
		Hits []openSearchHit `json:"hits"`
	} `json:"hits"`
}

type openSearchHit struct {
	Source openSearchMusicDocument `json:"_source"`
}

type openSearchMusicDocument struct {
	URI    string `json:"uri"`
	Type   string `json:"type"`
	Title  string `json:"title"`
	Artist string `json:"artist"`
	Album  string `json:"album"`
	Genre  string `json:"genre"`
}

// Search executes a Search Template request and converts hits to SearchResult.
func (s *OpenSearchExternalSearcher) Search(ctx context.Context, keyword string, resultTypes []SearchType, limit int) (*SearchResult, error) {
	l := limit
	if l <= 0 {
		l = 5
	}
	typeParam := searchTypesParam(resultTypes)

	ctx, span := otel.Tracer("owntone").Start(ctx, "owntone.ExternalSearch.Search")
	defer span.End()
	span.SetAttributes(
		attribute.String("search.engine", "opensearch"),
		attribute.String("search.index", s.config.Index),
		attribute.String("search.template_id", s.config.TemplateID),
		attribute.String("search.query_hash", hashInputText(keyword)),
		attribute.String("search.type", typeParam),
		attribute.Int("search.limit", l),
	)

	result, err := s.search(ctx, keyword, typeParam, l)
	if err != nil {
		recordExternalSearchFailure(span, externalSearchFallbackReason(err), err)
		return nil, err
	}
	hitCount := totalSearchResultCount(result)
	span.SetAttributes(attribute.Int("search.hit_count", hitCount))
	if hitCount == 0 {
		recordExternalSearchFailure(span, "no_hits", ErrExternalSearchNoHits)
		return nil, ErrExternalSearchNoHits
	}
	return result, nil
}

func (s *OpenSearchExternalSearcher) search(ctx context.Context, keyword string, typeParam string, limit int) (*SearchResult, error) {
	body, err := json.Marshal(openSearchTemplateRequest{
		ID: s.config.TemplateID,
		Params: openSearchTemplateParams{
			Query: keyword,
			Type:  typeParam,
			Size:  limit,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to encode external search request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.searchTemplateURL(), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to build external search request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send external search request: %w", err)
	}
	defer func() {
		_, _ = io.Copy(io.Discard, res.Body)
		_ = res.Body.Close()
	}()

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("unexpected external search status code: %d", res.StatusCode)
	}

	var response openSearchTemplateResponse
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode external search response: %w", err)
	}
	return openSearchResponseToSearchResult(response), nil
}

func (s *OpenSearchExternalSearcher) searchTemplateURL() string {
	base := strings.TrimRight(s.config.OpenSearchURL, "/")
	escapedIndex := pathEscape(s.config.Index)
	return base + "/" + escapedIndex + "/_search/template"
}

func pathEscape(value string) string {
	escaped := url.PathEscape(value)
	return strings.ReplaceAll(escaped, "%2F", "/")
}

func openSearchResponseToSearchResult(response openSearchTemplateResponse) *SearchResult {
	result := &SearchResult{}
	for _, hit := range response.Hits.Hits {
		doc := hit.Source
		if strings.TrimSpace(doc.URI) == "" {
			continue
		}
		item := SearchItem{
			Title:  doc.Title,
			URI:    doc.URI,
			Artist: doc.Artist,
		}
		switch doc.Type {
		case string(track):
			result.Tracks.Items = append(result.Tracks.Items, item)
		case string(artist):
			item.Title = ""
			item.Name = firstNonEmpty(doc.Artist, doc.Title)
			result.Artists.Items = append(result.Artists.Items, item)
		case string(album):
			item.Title = ""
			item.Name = firstNonEmpty(doc.Album, doc.Title)
			result.Albums.Items = append(result.Albums.Items, item)
		case string(genre):
			item.Title = ""
			item.Name = firstNonEmpty(doc.Genre, doc.Title)
			result.Genres.Items = append(result.Genres.Items, item)
		}
	}
	setExternalSearchTotals(result)
	return result
}


func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func setExternalSearchTotals(result *SearchResult) {
	for _, items := range []*Items{&result.Tracks, &result.Artists, &result.Albums, &result.Genres, &result.Playlists} {
		items.Total = len(items.Items)
		items.Offset = 0
		items.Limit = len(items.Items)
	}
}

func searchTypesParam(types []SearchType) string {
	if len(types) == 0 {
		return ""
	}
	values := make([]string, 0, len(types))
	for _, t := range types {
		values = append(values, string(t))
	}
	return strings.Join(values, ",")
}

func recordExternalSearchFailure(span trace.Span, reason string, err error) {
	span.SetAttributes(attribute.String("search.fallback_reason", reason))
	span.SetStatus(codes.Error, err.Error())
	span.RecordError(err)
}

func externalSearchFallbackReason(err error) string {
	if errors.Is(err, ErrExternalSearchNoHits) {
		return "no_hits"
	}
	msg := err.Error()
	switch {
	case strings.Contains(msg, "Client.Timeout"), strings.Contains(msg, "context deadline exceeded"):
		return "timeout"
	case strings.Contains(msg, "status code"):
		return "bad_status"
	case strings.Contains(msg, "decode"):
		return "invalid_json"
	case strings.Contains(msg, "send"):
		return "connection_error"
	default:
		return "error"
	}
}
