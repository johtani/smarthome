package otel

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"testing"

	"go.opentelemetry.io/otel"
	otellogglobal "go.opentelemetry.io/otel/log/global"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

type recordingLogExporter struct {
	mu       sync.Mutex
	records  []sdklog.Record
	shutdown bool
}

func (e *recordingLogExporter) Export(_ context.Context, records []sdklog.Record) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	for _, record := range records {
		e.records = append(e.records, record.Clone())
	}
	return nil
}

func (e *recordingLogExporter) ForceFlush(context.Context) error { return nil }

func (e *recordingLogExporter) Shutdown(context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.shutdown = true
	return nil
}

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
		spanID := span.SpanContext().SpanID().String()

		logger.InfoContext(ctx, "test message")

		var logRecord map[string]any
		if err := json.Unmarshal(buf.Bytes(), &logRecord); err != nil {
			t.Fatalf("failed to unmarshal log record: %v", err)
		}

		if logRecord["trace_id"] != traceID {
			t.Errorf("expected trace_id %s, got %v", traceID, logRecord["trace_id"])
		}
		if logRecord["span_id"] != spanID {
			t.Errorf("expected span_id %s, got %v", spanID, logRecord["span_id"])
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
		if _, ok := logRecord["span_id"]; ok {
			t.Error("expected no span_id in log record")
		}
	})
}

func TestLoggerProviderExportsSlogRecords(t *testing.T) {
	previousLogProvider := otellogglobal.GetLoggerProvider()
	defer otellogglobal.SetLoggerProvider(previousLogProvider)

	exporter := &recordingLogExporter{}
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String("test-service"),
	)
	loggerProvider := newLoggerProvider(exporter, res)
	otellogglobal.SetLoggerProvider(loggerProvider)

	var stderr bytes.Buffer
	logger := slog.New(NewLoggerHandler(slog.NewJSONHandler(&stderr, nil), slog.LevelInfo))

	traceProvider := trace.NewTracerProvider(trace.WithSampler(trace.AlwaysSample()))
	defer func() { _ = traceProvider.Shutdown(context.Background()) }()
	ctx, span := traceProvider.Tracer("test").Start(context.Background(), "test-span")
	traceID := span.SpanContext().TraceID()
	spanID := span.SpanContext().SpanID()

	logger.InfoContext(ctx, "with span", "key", "value")
	span.End()
	logger.InfoContext(context.Background(), "without span")
	logger.DebugContext(context.Background(), "filtered debug")

	if err := loggerProvider.ForceFlush(context.Background()); err != nil {
		t.Fatalf("failed to flush LoggerProvider: %v", err)
	}
	if err := loggerProvider.Shutdown(context.Background()); err != nil {
		t.Fatalf("failed to shutdown LoggerProvider: %v", err)
	}

	if got := bytes.Count(stderr.Bytes(), []byte("\n")); got != 2 {
		t.Errorf("expected existing handler to receive 2 records, got %d", got)
	}

	exporter.mu.Lock()
	defer exporter.mu.Unlock()
	if len(exporter.records) != 2 {
		t.Fatalf("expected 2 exported records, got %d", len(exporter.records))
	}
	if got := exporter.records[0].TraceID(); got != traceID {
		t.Errorf("expected trace ID %s, got %s", traceID, got)
	}
	if got := exporter.records[0].SpanID(); got != spanID {
		t.Errorf("expected span ID %s, got %s", spanID, got)
	}
	if exporter.records[1].TraceID().IsValid() || exporter.records[1].SpanID().IsValid() {
		t.Error("expected record without span context to have no trace or span ID")
	}
	if !exporter.shutdown {
		t.Error("expected exporter to be shut down")
	}
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    slog.Level
		wantErr bool
	}{
		{name: "default", want: slog.LevelInfo},
		{name: "debug", value: "debug", want: slog.LevelDebug},
		{name: "warn uppercase", value: "WARN", want: slog.LevelWarn},
		{name: "error", value: "error", want: slog.LevelError},
		{name: "invalid", value: "verbose", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseLogLevel(tt.value)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseLogLevel() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ParseLogLevel() = %v, want %v", got, tt.want)
			}
		})
	}
}
