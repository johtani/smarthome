package subcommand

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/johtani/smarthome/subcommand/action/llm"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type nlResolver interface {
	Path() string
	Prefix() string
	Resolve(ctx context.Context, text string, commandList string, promptVersion string) (llm.ResolvedCommand, error)
}

type legacyResolver struct {
	config llm.Config
}

func (r legacyResolver) Path() string {
	return "llm"
}

func (r legacyResolver) Prefix() string {
	return "LLM"
}

func (r legacyResolver) Resolve(ctx context.Context, text string, commandList string, promptVersion string) (llm.ResolvedCommand, error) {
	if strings.TrimSpace(r.config.Endpoint) == "" {
		return llm.ResolvedCommand{}, fmt.Errorf("llm endpoint is not configured")
	}
	return llm.NewClient(r.config).Resolve(ctx, text, commandList, promptVersion)
}

type dspyResolver struct {
	endpoint   string
	httpClient *http.Client
}

func newDSPyResolver(endpoint string, timeout time.Duration) dspyResolver {
	return dspyResolver{
		endpoint: strings.TrimSpace(endpoint),
		httpClient: &http.Client{
			Timeout:   timeout,
			Transport: otelhttp.NewTransport(http.DefaultTransport),
		},
	}
}

func (r dspyResolver) Path() string {
	return "dspy"
}

func (r dspyResolver) Prefix() string {
	return "DSPy"
}

func (r dspyResolver) Resolve(ctx context.Context, text string, commandList string, promptVersion string) (llm.ResolvedCommand, error) {
	if r.endpoint == "" {
		return llm.ResolvedCommand{}, fmt.Errorf("resolver.dspy_endpoint is not configured")
	}

	ctx, span := otel.Tracer("dspy").Start(ctx, "dspy.Resolve", trace.WithAttributes(
		attribute.String("resolver.path", "dspy"),
		attribute.String("resolver.prompt_version", promptVersion),
		attribute.String("dspy.endpoint", r.endpoint),
	))
	defer span.End()

	payload := map[string]string{
		"text":           text,
		"command_list":   commandList,
		"prompt_version": promptVersion,
	}
	reqBody, err := json.Marshal(payload)
	if err != nil {
		return llm.ResolvedCommand{}, fmt.Errorf("failed to marshal dspy payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.endpoint, bytes.NewReader(reqBody))
	if err != nil {
		return llm.ResolvedCommand{}, fmt.Errorf("failed to create dspy request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return llm.ResolvedCommand{}, fmt.Errorf("failed to send dspy request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return llm.ResolvedCommand{}, fmt.Errorf("failed to read dspy response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return llm.ResolvedCommand{}, fmt.Errorf("unexpected dspy status code: %d", resp.StatusCode)
	}

	var resolved llm.ResolvedCommand
	if err := json.Unmarshal(respBody, &resolved); err == nil {
		return resolved, nil
	}

	// Compatibility format: OpenAI-like response body
	var wrapped struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(respBody, &wrapped); err != nil {
		return llm.ResolvedCommand{}, fmt.Errorf("failed to decode dspy response: %w", err)
	}
	if len(wrapped.Choices) == 0 {
		return llm.ResolvedCommand{}, fmt.Errorf("no choices in dspy response")
	}
	if err := json.Unmarshal([]byte(wrapped.Choices[0].Message.Content), &resolved); err != nil {
		return llm.ResolvedCommand{}, fmt.Errorf("failed to decode dspy message content: %w", err)
	}
	return resolved, nil
}

func naturalLanguageResolvers(config Config) []nlResolver {
	legacyEnabled := strings.TrimSpace(config.LLM.Endpoint) != ""
	legacy := legacyResolver{config: config.LLM}
	if config.Resolver.Mode == ResolverModeDSPy {
		resolvers := []nlResolver{
			newDSPyResolver(config.Resolver.DSPyEndpoint, time.Duration(config.Resolver.DSPyTimeoutSeconds)*time.Second),
		}
		if legacyEnabled {
			resolvers = append(resolvers, legacy)
		}
		return resolvers
	}
	if legacyEnabled {
		return []nlResolver{legacy}
	}
	return nil
}
