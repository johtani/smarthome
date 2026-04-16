package llm

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestClient_Resolve(t *testing.T) {
	exporter, shutdown := setupTestTracerProvider(t)
	defer shutdown()
	var requestBody string

	mockResponse := struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}{
		Choices: []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		}{
			{
				Message: struct {
					Content string `json:"content"`
				}{
					Content: `{"command": "light on", "args": "", "thought": "user wants to turn on the light"}`,
				},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST request, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected application/json content-type, got %s", r.Header.Get("Content-Type"))
		}
		body, _ := io.ReadAll(r.Body)
		requestBody = string(body)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	config := Config{
		Endpoint: server.URL,
		Model:    "test-model",
	}
	client := NewClient(config)

	resolved, err := client.Resolve(t.Context(), "電気をつけて", "light on: turn on the light", "v1")
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	if resolved.Command != "light on" {
		t.Errorf("expected command 'light on', got '%s'", resolved.Command)
	}
	if resolved.Thought == "" {
		t.Error("expected thought to be non-empty")
	}
	if !strings.Contains(requestBody, "コマンドの args 指定に必ず従ってください") {
		t.Errorf("expected prompt rule in request body, got %s", requestBody)
	}

	span := findResolveSpan(t, exporter)
	attrs := attrsAsMap(span.Attributes)

	if attrs["llm.endpoint"] != server.URL {
		t.Errorf("expected llm.endpoint %q, got %q", server.URL, attrs["llm.endpoint"])
	}
	if attrs["llm.request_body"] == "" {
		t.Error("expected llm.request_body to be set")
	}
	if attrs["llm.response_body"] == "" {
		t.Error("expected llm.response_body to be set")
	}
	if attrs["llm.response_status_code"] != "200" {
		t.Errorf("expected llm.response_status_code 200, got %q", attrs["llm.response_status_code"])
	}
	if attrs["resolver.prompt_version"] != "v1" {
		t.Errorf("expected resolver.prompt_version v1, got %q", attrs["resolver.prompt_version"])
	}
}

func TestClient_ResolveErrorSetsResponseTraceAttributes(t *testing.T) {
	exporter, shutdown := setupTestTracerProvider(t)
	defer shutdown()

	const errorBody = `{"error":"bad request"}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, errorBody, http.StatusBadRequest)
	}))
	defer server.Close()

	config := Config{
		Endpoint: server.URL,
		Model:    "test-model",
	}
	client := NewClient(config)

	_, err := client.Resolve(t.Context(), "電気をつけて", "light on: turn on the light", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	span := findResolveSpan(t, exporter)
	attrs := attrsAsMap(span.Attributes)

	if attrs["llm.response_status_code"] != "400" {
		t.Errorf("expected llm.response_status_code 400, got %q", attrs["llm.response_status_code"])
	}
	if !strings.Contains(attrs["llm.response_body"], "bad request") {
		t.Errorf("expected llm.response_body to contain error payload, got %q", attrs["llm.response_body"])
	}
}

func TestClient_ResolveTruncatesLargeResponseBodyInTrace(t *testing.T) {
	exporter, shutdown := setupTestTracerProvider(t)
	defer shutdown()

	longBody := strings.Repeat("x", maxTraceBodyLength+10)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(longBody))
	}))
	defer server.Close()

	config := Config{
		Endpoint: server.URL,
		Model:    "test-model",
	}
	client := NewClient(config)

	_, err := client.Resolve(t.Context(), "電気をつけて", "light on: turn on the light", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	span := findResolveSpan(t, exporter)
	attrs := attrsAsMap(span.Attributes)
	got := attrs["llm.response_body"]

	if !strings.HasSuffix(got, "...truncated") {
		t.Errorf("expected truncated suffix, got %q", got)
	}
	if len(got) != maxTraceBodyLength+len("...truncated") {
		t.Errorf("expected truncated length %d, got %d", maxTraceBodyLength+len("...truncated"), len(got))
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				Endpoint: "http://localhost:8080",
				Model:    "gpt-4o",
			},
			wantErr: false,
		},
		{
			name:    "disabled config (both empty)",
			config:  Config{},
			wantErr: false,
		},
		{
			name: "missing model when endpoint is set",
			config: Config{
				Endpoint: "http://localhost:8080",
			},
			wantErr: true,
		},
		{
			name: "missing endpoint when model is set",
			config: Config{
				Model: "gpt-4o",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.config.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func setupTestTracerProvider(t *testing.T) (*tracetest.InMemoryExporter, func()) {
	t.Helper()
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithSyncer(exporter),
	)

	prev := otel.GetTracerProvider()
	otel.SetTracerProvider(tp)

	return exporter, func() {
		_ = tp.Shutdown(t.Context())
		otel.SetTracerProvider(prev)
	}
}

func findResolveSpan(t *testing.T, exporter *tracetest.InMemoryExporter) tracetest.SpanStub {
	t.Helper()
	spans := exporter.GetSpans()
	for _, s := range spans {
		if s.Name == "llm.Resolve" {
			return s
		}
	}
	t.Fatalf("expected llm.Resolve span, got %d spans", len(spans))
	return tracetest.SpanStub{}
}

func attrsAsMap(attrs []attribute.KeyValue) map[string]string {
	m := make(map[string]string, len(attrs))
	for _, kv := range attrs {
		m[string(kv.Key)] = kv.Value.Emit()
	}
	return m
}
