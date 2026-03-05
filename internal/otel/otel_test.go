package otel

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestTracingHandler(t *testing.T) {
	// Setup a buffer to capture logs
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, nil)
	tracingHandler := NewTracingHandler(handler)
	logger := slog.New(tracingHandler)

	// Setup OpenTelemetry
	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(trace.WithBatcher(exporter))
	otel.SetTracerProvider(tp)
	defer func() { _ = tp.Shutdown(context.Background()) }()

	tracer := tp.Tracer("test")

	t.Run("with trace id", func(t *testing.T) {
		buf.Reset()
		ctx, span := tracer.Start(context.Background(), "test-span")
		defer span.End()

		traceID := span.SpanContext().TraceID().String()

		logger.InfoContext(ctx, "test message")

		var logRecord map[string]any
		if err := json.Unmarshal(buf.Bytes(), &logRecord); err != nil {
			t.Fatalf("failed to unmarshal log record: %v", err)
		}

		if logRecord["trace_id"] != traceID {
			t.Errorf("expected trace_id %s, got %v", traceID, logRecord["trace_id"])
		}
	})

	t.Run("without trace id", func(t *testing.T) {
		buf.Reset()
		logger.InfoContext(context.Background(), "test message")

		var logRecord map[string]any
		if err := json.Unmarshal(buf.Bytes(), &logRecord); err != nil {
			t.Fatalf("failed to unmarshal log record: %v", err)
		}

		if _, ok := logRecord["trace_id"]; ok {
			t.Error("expected no trace_id in log record")
		}
	})
}
