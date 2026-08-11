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

func TestCommandsFindTracing_DSPyPath(t *testing.T) {
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
		_, _ = w.Write([]byte(`{"command":"light on","args":"","thought":"resolved by DSPy"}`))
	}))
	defer server.Close()

	cmds := Commands{
		Definitions: []Definition{
			{Name: "light on", Description: "turn on light", Factory: NewDummySubcommand},
		},
	}
	config := Config{
		Resolver: ResolverConfig{
			Mode:               ResolverModeDSPy,
			DSPyEndpoint:       server.URL,
			DSPyTimeoutSeconds: 3,
			PromptVersion:      "v2",
		},
	}

	_, _, _, err := cmds.Find(t.Context(), config, "あかりをつけて")
	if err != nil {
		t.Fatalf("Find failed: %v", err)
	}

	span := findSpanByName(t, exporter, "Commands.Find")
	attrs := toAttrMap(span.Attributes)
	if attrs["resolver.path"] != "dspy" {
		t.Fatalf("expected resolver.path dspy, got %q", attrs["resolver.path"])
	}
	if attrs["resolver.resolved_command"] != "light on" {
		t.Fatalf("expected resolver.resolved_command light on, got %q", attrs["resolver.resolved_command"])
	}

	dspySpan := findSpanByName(t, exporter, "dspy.Resolve")
	dspyAttrs := toAttrMap(dspySpan.Attributes)
	if dspyAttrs["resolver.prompt_version"] != "v2" {
		t.Fatalf("expected resolver.prompt_version v2, got %q", dspyAttrs["resolver.prompt_version"])
	}
	if dspyAttrs["dspy.request.text"] != "あかりをつけて" {
		t.Fatalf("expected dspy.request.text あかりをつけて, got %q", dspyAttrs["dspy.request.text"])
	}
	if dspyAttrs["dspy.request.prompt_version"] != "v2" {
		t.Fatalf("expected dspy.request.prompt_version v2, got %q", dspyAttrs["dspy.request.prompt_version"])
	}
	if dspyAttrs["dspy.request.command_count"] != "2" {
		t.Fatalf("expected dspy.request.command_count 2, got %q", dspyAttrs["dspy.request.command_count"])
	}
	if dspyAttrs["dspy.response_body"] != `{"command":"light on","args":"","thought":"resolved by DSPy"}` {
		t.Fatalf("expected dspy.response_body %s, got %q", `{"command":"light on","args":"","thought":"resolved by DSPy"}`, dspyAttrs["dspy.response_body"])
	}
}

func TestCommandsFindTracing_DSPyMusicCorrection(t *testing.T) {
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
		_, _ = w.Write([]byte(`{"command":"start music","args":"artist","thought":"artist playback"}`))
	}))
	defer server.Close()

	cmds := Commands{Definitions: []Definition{
		NewStartMusicCmdDefinition(),
		NewSearchAndPlayMusicCmdDefinition(),
	}}
	config := Config{Resolver: ResolverConfig{
		Mode:               ResolverModeDSPy,
		DSPyEndpoint:       server.URL,
		DSPyTimeoutSeconds: 3,
	}}

	def, _, _, err := cmds.Find(t.Context(), config, "B'zの曲をかけて")
	if err != nil {
		t.Fatalf("Find failed: %v", err)
	}
	if def.Name != SearchAndPlayMusicCmd {
		t.Fatalf("command = %q, want %q", def.Name, SearchAndPlayMusicCmd)
	}

	span := findSpanByName(t, exporter, "Commands.Find")
	attrs := toAttrMap(span.Attributes)
	if attrs["resolver.initial_command"] != StartMusicCmd {
		t.Fatalf("initial command = %q, want %q", attrs["resolver.initial_command"], StartMusicCmd)
	}
	if attrs["resolver.initial_args_kind"] != "mode" {
		t.Fatalf("initial args kind = %q, want mode", attrs["resolver.initial_args_kind"])
	}
	if attrs["resolver.command_corrected"] != "true" {
		t.Fatalf("command corrected = %q, want true", attrs["resolver.command_corrected"])
	}
	if attrs["resolver.correction_reason"] != specifiedMusicTargetCorrectionReason {
		t.Fatalf("correction reason = %q, want %q", attrs["resolver.correction_reason"], specifiedMusicTargetCorrectionReason)
	}

	var decisionAttrs map[string]string
	for _, event := range span.Events {
		if event.Name == "resolver.decision" {
			decisionAttrs = toAttrMap(event.Attributes)
			break
		}
	}
	if decisionAttrs == nil {
		t.Fatal("resolver.decision event not found")
	}
	if decisionAttrs["resolver.initial_command"] != StartMusicCmd {
		t.Fatalf("event initial command = %q, want %q", decisionAttrs["resolver.initial_command"], StartMusicCmd)
	}
	if decisionAttrs["resolver.command_corrected"] != "true" {
		t.Fatalf("event command corrected = %q, want true", decisionAttrs["resolver.command_corrected"])
	}
}

func TestCommandsFindTracing_DSPyInvalidArgs(t *testing.T) {
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
		_, _ = w.Write([]byte(`{"command":"start music","args":"mode:random","thought":"invalid mode"}`))
	}))
	defer server.Close()

	cmds := Commands{Definitions: []Definition{NewStartMusicCmdDefinition()}}
	config := Config{Resolver: ResolverConfig{
		Mode:               ResolverModeDSPy,
		DSPyEndpoint:       server.URL,
		DSPyTimeoutSeconds: 3,
	}}

	if _, _, _, err := cmds.Find(t.Context(), config, "音楽をかけて"); err == nil {
		t.Fatal("Find succeeded with invalid resolver arguments")
	}

	span := findSpanByName(t, exporter, "Commands.Find")
	attrs := toAttrMap(span.Attributes)
	if attrs["resolver.args.valid"] != "false" {
		t.Fatalf("resolver.args.valid = %q, want false", attrs["resolver.args.valid"])
	}
	if attrs["resolver.args.validation_reason"] != argsValidationReasonInvalidPrefix {
		t.Fatalf(
			"resolver.args.validation_reason = %q, want %q",
			attrs["resolver.args.validation_reason"],
			argsValidationReasonInvalidPrefix,
		)
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
