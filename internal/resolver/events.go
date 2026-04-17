package resolver

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// EventSchemaVersion is the fixed schema version for resolver telemetry events.
const EventSchemaVersion = "v1"

const (
	eventNameDecision  = "resolver.decision"
	eventNameExecution = "resolver.execution"
	eventNameFeedback  = "resolver.feedback"
)

// DecisionRecord represents a fixed schema for resolver decision events.
type DecisionRecord struct {
	InputTextHash     string
	ResolverPath      string
	ResolverMode      string
	LLMModel          string
	ResolvedCommand   string
	ResolvedArgs      string
	DidYouMeanCommand string
}

// ExecutionRecord represents a fixed schema for execution result events.
type ExecutionRecord struct {
	ExecutionStatus  string
	ResolvedCommand  string
	ResolvedArgs     string
	ResolverPathHint string
}

// FeedbackRecord represents a fixed schema for user feedback events.
type FeedbackRecord struct {
	FeedbackLabel      string
	FeedbackCorrection string
	ResolvedCommand    string
	ResolvedArgs       string
	ResolverPathHint   string
}

// RecordDecision records a resolver decision event with a fixed schema.
func RecordDecision(ctx context.Context, record DecisionRecord) {
	attrs := append(baseAttrsFromContext(ctx),
		attribute.String("resolver.schema_version", EventSchemaVersion),
		attribute.String("resolver.input_text_hash", record.InputTextHash),
		attribute.String("resolver.path", record.ResolverPath),
		attribute.String("resolver.mode", record.ResolverMode),
		attribute.String("llm.model", record.LLMModel),
		attribute.String("resolver.resolved_command", record.ResolvedCommand),
		attribute.String("resolver.resolved_args", record.ResolvedArgs),
		attribute.String("resolver.did_you_mean_command", record.DidYouMeanCommand),
	)
	span := trace.SpanFromContext(ctx)
	span.AddEvent(eventNameDecision, trace.WithAttributes(attrs...))

	slog.InfoContext(
		ctx,
		"resolver decision recorded",
		"resolver.schema_version", EventSchemaVersion,
		"resolver.input_text_hash", record.InputTextHash,
		"resolver.path", record.ResolverPath,
		"resolver.mode", record.ResolverMode,
		"llm.model", record.LLMModel,
		"resolver.resolved_command", record.ResolvedCommand,
		"resolver.resolved_args", record.ResolvedArgs,
		"resolver.did_you_mean_command", record.DidYouMeanCommand,
	)
}

// RecordExecution records a resolver execution event with a fixed schema.
func RecordExecution(ctx context.Context, record ExecutionRecord) {
	attrs := append(baseAttrsFromContext(ctx),
		attribute.String("resolver.schema_version", EventSchemaVersion),
		attribute.String("resolver.execution_status", record.ExecutionStatus),
		attribute.String("resolver.resolved_command", record.ResolvedCommand),
		attribute.String("resolver.resolved_args", record.ResolvedArgs),
		attribute.String("resolver.path_hint", record.ResolverPathHint),
	)
	span := trace.SpanFromContext(ctx)
	span.AddEvent(eventNameExecution, trace.WithAttributes(attrs...))

	slog.InfoContext(
		ctx,
		"resolver execution recorded",
		"resolver.schema_version", EventSchemaVersion,
		"resolver.execution_status", record.ExecutionStatus,
		"resolver.resolved_command", record.ResolvedCommand,
		"resolver.resolved_args", record.ResolvedArgs,
		"resolver.path_hint", record.ResolverPathHint,
	)
}

// RecordFeedback records a user feedback event with a fixed schema.
func RecordFeedback(ctx context.Context, record FeedbackRecord) {
	attrs := append(baseAttrsFromContext(ctx),
		attribute.String("resolver.schema_version", EventSchemaVersion),
		attribute.String("feedback.label", record.FeedbackLabel),
		attribute.String("feedback.correction", record.FeedbackCorrection),
		attribute.String("resolver.resolved_command", record.ResolvedCommand),
		attribute.String("resolver.resolved_args", record.ResolvedArgs),
		attribute.String("resolver.path_hint", record.ResolverPathHint),
	)
	span := trace.SpanFromContext(ctx)
	span.AddEvent(eventNameFeedback, trace.WithAttributes(attrs...))

	slog.InfoContext(
		ctx,
		"resolver feedback recorded",
		"resolver.schema_version", EventSchemaVersion,
		"feedback.label", record.FeedbackLabel,
		"feedback.correction", record.FeedbackCorrection,
		"resolver.resolved_command", record.ResolvedCommand,
		"resolver.resolved_args", record.ResolvedArgs,
		"resolver.path_hint", record.ResolverPathHint,
	)
}

func baseAttrsFromContext(ctx context.Context) []attribute.KeyValue {
	attrs := make([]attribute.KeyValue, 0, 2)
	if requestID, ok := RequestIDFromContext(ctx); ok {
		attrs = append(attrs, attribute.String("resolver.request_id", requestID))
	}
	if channel, ok := ChannelFromContext(ctx); ok {
		attrs = append(attrs, attribute.String("resolver.channel", channel))
	}
	return attrs
}
