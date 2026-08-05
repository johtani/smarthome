/*
Package otel provides OpenTelemetry initialization and configuration for the application.
*/
package otel

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	otellogglobal "go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

// TracingHandler is a slog.Handler that adds trace and span IDs to log records from context.
type TracingHandler struct {
	slog.Handler
}

// Handle adds trace_id and span_id to the record if they exist in the context.
func (h *TracingHandler) Handle(ctx context.Context, r slog.Record) error {
	if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
		r.AddAttrs(
			slog.String("trace_id", sc.TraceID().String()),
			slog.String("span_id", sc.SpanID().String()),
		)
	}
	return h.Handler.Handle(ctx, r)
}

// NewTracingHandler creates a new TracingHandler.
func NewTracingHandler(h slog.Handler) *TracingHandler {
	return &TracingHandler{Handler: h}
}

// fanoutHandler sends each log record to all configured handlers.
type fanoutHandler struct {
	handlers []slog.Handler
	minLevel slog.Leveler
}

func (h *fanoutHandler) Enabled(ctx context.Context, level slog.Level) bool {
	if level < h.minLevel.Level() {
		return false
	}
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (h *fanoutHandler) Handle(ctx context.Context, record slog.Record) error {
	if record.Level < h.minLevel.Level() {
		return nil
	}
	var errs []error
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, record.Level) {
			errs = append(errs, handler.Handle(ctx, record.Clone()))
		}
	}
	return errors.Join(errs...)
}

func (h *fanoutHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		handlers[i] = handler.WithAttrs(attrs)
	}
	return &fanoutHandler{handlers: handlers, minLevel: h.minLevel}
}

func (h *fanoutHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		handlers[i] = handler.WithGroup(name)
	}
	return &fanoutHandler{handlers: handlers, minLevel: h.minLevel}
}

// NewLoggerHandler creates a handler that preserves the existing log output and
// also emits records through the globally configured OpenTelemetry LoggerProvider.
func NewLoggerHandler(existing slog.Handler, minLevel slog.Leveler) slog.Handler {
	return &fanoutHandler{handlers: []slog.Handler{
		existing,
		otelslog.NewHandler("github.com/johtani/smarthome"),
	}, minLevel: minLevel}
}

// ParseLogLevel parses an application log level. An empty value defaults to info.
func ParseLogLevel(value string) (slog.Level, error) {
	if strings.TrimSpace(value) == "" {
		return slog.LevelInfo, nil
	}
	var level slog.Level
	if err := level.UnmarshalText([]byte(value)); err != nil {
		return 0, fmt.Errorf("invalid log level %q: %w", value, err)
	}
	return level, nil
}

func newLoggerProvider(exporter sdklog.Exporter, res *resource.Resource) *sdklog.LoggerProvider {
	return sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(exporter)),
		sdklog.WithResource(res),
	)
}

// SetupOTEL OpenTelemetryの設定を行い、クリーンアップ関数を返します。
func SetupOTEL(ctx context.Context, serviceName string) (func(context.Context) error, error) {
	var shutdownFuncs []func(context.Context) error

	// クリーンアップ関数をまとめる
	shutdown := func(ctx context.Context) error {
		var errs []error
		for _, fn := range shutdownFuncs {
			if e := fn(ctx); e != nil {
				errs = append(errs, e)
			}
		}
		return errors.Join(errs...)
	}

	// Propagatorの設定 (TraceParent等を伝播させるため)
	prop := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
	otel.SetTextMapPropagator(prop)

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Trace Exporter (OTLP/HTTP)
	// 環境変数 (OTEL_EXPORTER_OTLP_ENDPOINT等) で設定可能
	traceExporter, err := otlptracehttp.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	// TracerProvider
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter, sdktrace.WithBatchTimeout(5*time.Second)),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tracerProvider)
	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)

	// Metric Exporter (OTLP/HTTP)
	metricExporter, err := otlpmetrichttp.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create metric exporter: %w", err)
	}

	// MeterProvider
	meterProvider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(metricExporter, metric.WithInterval(5*time.Second))),
		metric.WithResource(res),
	)
	otel.SetMeterProvider(meterProvider)
	shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)

	// Log Exporter and LoggerProvider (OTLP/HTTP)
	logExporter, err := otlploghttp.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create log exporter: %w", err)
	}
	loggerProvider := newLoggerProvider(logExporter, res)
	otellogglobal.SetLoggerProvider(loggerProvider)
	shutdownFuncs = append(shutdownFuncs, func(ctx context.Context) error {
		return errors.Join(loggerProvider.ForceFlush(ctx), loggerProvider.Shutdown(ctx))
	})

	return shutdown, nil
}
