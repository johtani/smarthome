package owntone

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// MusicIntentResolver resolves free text into a structured music intent.
type MusicIntentResolver interface {
	Path() string
	Resolve(ctx context.Context, text string) (MusicIntent, error)
}

type httpMusicIntentResolver struct {
	endpoint   string
	httpClient *http.Client
}

// NewHTTPMusicIntentResolver creates an HTTP-based music intent resolver.
func NewHTTPMusicIntentResolver(endpoint string, timeout time.Duration) MusicIntentResolver {
	trimmedEndpoint := strings.TrimSpace(endpoint)
	if trimmedEndpoint == "" {
		return nil
	}
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	return httpMusicIntentResolver{
		endpoint: trimmedEndpoint,
		httpClient: &http.Client{
			Timeout:   timeout,
			Transport: otelhttp.NewTransport(http.DefaultTransport),
		},
	}
}

func (r httpMusicIntentResolver) Path() string {
	return "music_intent_http"
}

func (r httpMusicIntentResolver) Resolve(ctx context.Context, text string) (MusicIntent, error) {
	payload, err := json.Marshal(map[string]string{"text": strings.TrimSpace(text)})
	if err != nil {
		return MusicIntent{}, fmt.Errorf("failed to marshal music intent payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.endpoint, bytes.NewReader(payload))
	if err != nil {
		return MusicIntent{}, fmt.Errorf("failed to build music intent request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return MusicIntent{}, fmt.Errorf("failed to send music intent request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return MusicIntent{}, fmt.Errorf("failed to read music intent response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return MusicIntent{}, fmt.Errorf("unexpected music intent status code: %d", resp.StatusCode)
	}

	var wrapped struct {
		MusicIntent MusicIntent `json:"music_intent"`
		Model       string      `json:"model"`
		Reason      string      `json:"reason"`
	}
	if err := json.Unmarshal(body, &wrapped); err == nil && !wrapped.MusicIntent.IsEmpty() {
		if strings.TrimSpace(wrapped.MusicIntent.Model) == "" {
			wrapped.MusicIntent.Model = strings.TrimSpace(wrapped.Model)
		}
		if strings.TrimSpace(wrapped.MusicIntent.Reason) == "" {
			wrapped.MusicIntent.Reason = strings.TrimSpace(wrapped.Reason)
		}
		return wrapped.MusicIntent, nil
	}

	var intent MusicIntent
	if err := json.Unmarshal(body, &intent); err != nil {
		return MusicIntent{}, fmt.Errorf("failed to decode music intent response: %w", err)
	}
	return intent, nil
}
