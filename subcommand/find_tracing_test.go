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
	decisionAttrs := findEventAttrs(t, span, "resolver.decision")
	if decisionAttrs["llm.model"] != "test-model" {
		t.Fatalf("expected decision llm.model test-model, got %q", decisionAttrs["llm.model"])
	}
	if decisionAttrs["resolver.prompt_version"] != "v1" {
		t.Fatalf("expected decision prompt version v1, got %q", decisionAttrs["resolver.prompt_version"])
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
		_, _ = w.Write([]byte(`{"command":"light on","args":"","thought":"resolved by DSPy","model":"lfm2.5-2.6B","prompt_version":"v2","artifact_version":"artifact-v3","dataset_version":"dataset-v4"}`))
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
	if dspyAttrs["dspy.response_body"] == "" {
		t.Fatal("expected dspy.response_body")
	}
	decisionAttrs := findEventAttrs(t, span, "resolver.decision")
	if decisionAttrs["llm.model"] != "lfm2.5-2.6B" {
		t.Fatalf("expected decision llm.model lfm2.5-2.6B, got %q", decisionAttrs["llm.model"])
	}
	if decisionAttrs["resolver.prompt_version"] != "v2" {
		t.Fatalf("expected decision prompt version v2, got %q", decisionAttrs["resolver.prompt_version"])
	}
	if decisionAttrs["resolver.artifact_version"] != "artifact-v3" {
		t.Fatalf("expected artifact version artifact-v3, got %q", decisionAttrs["resolver.artifact_version"])
	}
	if decisionAttrs["resolver.dataset_version"] != "dataset-v4" {
		t.Fatalf("expected dataset version dataset-v4, got %q", decisionAttrs["resolver.dataset_version"])
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

func findEventAttrs(t *testing.T, span tracetest.SpanStub, name string) map[string]string {
	t.Helper()
	for _, event := range span.Events {
		if event.Name == name {
			return toAttrMap(event.Attributes)
		}
	}
	t.Fatalf("event %q not found", name)
	return nil
}

func toAttrMap(attrs []attribute.KeyValue) map[string]string {
	m := make(map[string]string, len(attrs))
	for _, kv := range attrs {
		m[string(kv.Key)] = kv.Value.Emit()
	}
	return m
}
