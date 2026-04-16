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
		InputTextHash:   "abc",
		ResolverPath:    "llm",
		ResolverMode:    "legacy",
		LLMModel:        "gpt-4o",
		ResolvedCommand: "light on",
		ResolvedArgs:    "",
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
