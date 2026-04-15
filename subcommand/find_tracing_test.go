package subcommand

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/johtani/smarthome/internal/resolver"
	"github.com/johtani/smarthome/subcommand/action/llm"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestCommandsFindTracing_ExactMatch(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithSyncer(exporter),
	)
	defer func() { _ = tp.Shutdown(t.Context()) }()

	prev := otel.GetTracerProvider()
	otel.SetTracerProvider(tp)
	defer otel.SetTracerProvider(prev)

	cmds := Commands{
		Definitions: []Definition{
			{Name: "light on", Description: "turn on light", Factory: NewDummySubcommand},
		},
	}
	config := Config{
		Resolver: ResolverConfig{Mode: ResolverModeLegacy},
	}
	ctx := resolver.WithRequestID(t.Context(), "req-1")
	ctx = resolver.WithChannel(ctx, "slack_mention")

	_, _, _, err := cmds.Find(ctx, config, "light on")
	if err != nil {
		t.Fatalf("Find failed: %v", err)
	}

	span := findSpanByName(t, exporter, "Commands.Find")
	attrs := toAttrMap(span.Attributes)
	if attrs["resolver.path"] != "exact_match" {
		t.Fatalf("expected resolver.path exact_match, got %q", attrs["resolver.path"])
	}
	if attrs["resolver.request_id"] != "req-1" {
		t.Fatalf("expected resolver.request_id req-1, got %q", attrs["resolver.request_id"])
	}
	if attrs["resolver.channel"] != "slack_mention" {
		t.Fatalf("expected resolver.channel slack_mention, got %q", attrs["resolver.channel"])
	}
}

func TestCommandsFindTracing_LLMPath(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithSyncer(exporter),
	)
	defer func() { _ = tp.Shutdown(t.Context()) }()

	prev := otel.GetTracerProvider()
	otel.SetTracerProvider(tp)
	defer otel.SetTracerProvider(prev)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := struct {
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
						Content: `{"command": "light on", "args": "", "thought": "resolved by LLM"}`,
					},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cmds := Commands{
		Definitions: []Definition{
			{Name: "light on", Description: "turn on light", Factory: NewDummySubcommand},
		},
	}
	config := Config{
		LLM: llm.Config{
			Endpoint: server.URL,
			Model:    "test-model",
		},
		Resolver: ResolverConfig{
			Mode:          ResolverModeLegacy,
			PromptVersion: "v1",
		},
	}

	_, _, _, err := cmds.Find(t.Context(), config, "あかりをつけて")
	if err != nil {
		t.Fatalf("Find failed: %v", err)
	}

	span := findSpanByName(t, exporter, "Commands.Find")
	attrs := toAttrMap(span.Attributes)
	if attrs["resolver.path"] != "llm" {
		t.Fatalf("expected resolver.path llm, got %q", attrs["resolver.path"])
	}
	if attrs["resolver.resolved_command"] != "light on" {
		t.Fatalf("expected resolver.resolved_command light on, got %q", attrs["resolver.resolved_command"])
	}

	llmSpan := findSpanByName(t, exporter, "llm.Resolve")
	llmAttrs := toAttrMap(llmSpan.Attributes)
	if llmAttrs["resolver.prompt_version"] != "v1" {
		t.Fatalf("expected resolver.prompt_version v1, got %q", llmAttrs["resolver.prompt_version"])
	}
}

func findSpanByName(t *testing.T, exporter *tracetest.InMemoryExporter, name string) tracetest.SpanStub {
	t.Helper()
	spans := exporter.GetSpans()
	for _, s := range spans {
		if s.Name == name {
			return s
		}
	}
	t.Fatalf("expected span %q, got %d spans", name, len(spans))
	return tracetest.SpanStub{}
}

func toAttrMap(attrs []attribute.KeyValue) map[string]string {
	m := make(map[string]string, len(attrs))
	for _, kv := range attrs {
		m[string(kv.Key)] = kv.Value.Emit()
	}
	return m
}
