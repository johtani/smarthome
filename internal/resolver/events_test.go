package resolver

import (
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestRecordDecisionAddsEvent(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithSyncer(exporter),
	)
	defer func() { _ = tp.Shutdown(t.Context()) }()

	prev := otel.GetTracerProvider()
	otel.SetTracerProvider(tp)
	defer otel.SetTracerProvider(prev)

	ctx := WithRequestID(t.Context(), "req-1")
	ctx = WithChannel(ctx, "slack_mention")
	ctx, span := otel.Tracer("test").Start(ctx, "span")
	RecordDecision(ctx, DecisionRecord{
		InputTextHash:    "abc",
		ResolverPath:     "llm",
		ResolverMode:     "legacy",
		LLMModel:         "gpt-4o",
		PromptVersion:    "prompt-v2",
		ArtifactVersion:  "artifact-v3",
		DatasetVersion:   "dataset-v4",
		InitialCommand:   "start music",
		InitialArgsKind:  "mode",
		CommandCorrected: true,
		CorrectionReason: "specified_music_target",
		ResolvedCommand:  "light on",
		ResolvedArgs:     "",
	})
	span.End()

	recorded := findSpan(t, exporter, "span")
	if len(recorded.Events) == 0 {
		t.Fatal("expected at least one event")
	}
	ev := recorded.Events[0]
	if ev.Name != "resolver.decision" {
		t.Fatalf("expected event resolver.decision, got %q", ev.Name)
	}
	attrs := toMap(ev.Attributes)
	if attrs["resolver.request_id"] != "req-1" {
		t.Fatalf("expected resolver.request_id req-1, got %q", attrs["resolver.request_id"])
	}
	if attrs["resolver.schema_version"] != EventSchemaVersion {
		t.Fatalf("expected schema version %q, got %q", EventSchemaVersion, attrs["resolver.schema_version"])
	}
	if attrs["resolver.initial_command"] != "start music" {
		t.Fatalf("expected initial command start music, got %q", attrs["resolver.initial_command"])
	}
	if attrs["resolver.initial_args_kind"] != "mode" {
		t.Fatalf("expected initial args kind mode, got %q", attrs["resolver.initial_args_kind"])
	}
	if attrs["resolver.command_corrected"] != "true" {
		t.Fatalf("expected command corrected true, got %q", attrs["resolver.command_corrected"])
	}
	if attrs["resolver.correction_reason"] != "specified_music_target" {
		t.Fatalf("expected correction reason specified_music_target, got %q", attrs["resolver.correction_reason"])
	}
	if attrs["llm.model"] != "gpt-4o" {
		t.Fatalf("expected llm.model gpt-4o, got %q", attrs["llm.model"])
	}
	if attrs["resolver.prompt_version"] != "prompt-v2" {
		t.Fatalf("expected prompt version prompt-v2, got %q", attrs["resolver.prompt_version"])
	}
	if attrs["resolver.artifact_version"] != "artifact-v3" {
		t.Fatalf("expected artifact version artifact-v3, got %q", attrs["resolver.artifact_version"])
	}
	if attrs["resolver.dataset_version"] != "dataset-v4" {
		t.Fatalf("expected dataset version dataset-v4, got %q", attrs["resolver.dataset_version"])
	}
}

func TestRecordExecutionAndFeedbackAddsEvents(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithSyncer(exporter),
	)
	defer func() { _ = tp.Shutdown(t.Context()) }()

	prev := otel.GetTracerProvider()
	otel.SetTracerProvider(tp)
	defer otel.SetTracerProvider(prev)

	ctx := WithRequestID(t.Context(), "req-2")
	ctx = WithChannel(ctx, "slack_feedback")
	ctx, span := otel.Tracer("test").Start(ctx, "span2")
	RecordExecution(ctx, ExecutionRecord{
		ExecutionStatus: "success",
		ResolvedCommand: "light on",
	})
	RecordFeedback(ctx, FeedbackRecord{
		FeedbackLabel:      "incorrect",
		FeedbackCorrection: "light off",
		ResolvedCommand:    "light on",
		ResolverPathHint:   "unresolved",
	})
	span.End()

	recorded := findSpan(t, exporter, "span2")
	if len(recorded.Events) < 2 {
		t.Fatalf("expected at least 2 events, got %d", len(recorded.Events))
	}
	if recorded.Events[0].Name != "resolver.execution" {
		t.Fatalf("expected first event resolver.execution, got %q", recorded.Events[0].Name)
	}
	if recorded.Events[1].Name != "resolver.feedback" {
		t.Fatalf("expected second event resolver.feedback, got %q", recorded.Events[1].Name)
	}
	attrs := toMap(recorded.Events[1].Attributes)
	if attrs["resolver.path_hint"] != "unresolved" {
		t.Fatalf("expected resolver.path_hint unresolved, got %q", attrs["resolver.path_hint"])
	}
}

func findSpan(t *testing.T, exporter *tracetest.InMemoryExporter, name string) tracetest.SpanStub {
	t.Helper()
	for _, span := range exporter.GetSpans() {
		if span.Name == name {
			return span
		}
	}
	t.Fatalf("span %q not found", name)
	return tracetest.SpanStub{}
}

func toMap(attrs []attribute.KeyValue) map[string]string {
	m := make(map[string]string, len(attrs))
	for _, kv := range attrs {
		m[string(kv.Key)] = kv.Value.Emit()
	}
	return m
}
