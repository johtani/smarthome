/*
Package action defines the interface for smart home actions.
Actions are the smallest units of work, such as calling an API or controlling a device.
*/
package action

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Action is an interface for executing a single smart home action.
type Action interface {
	Run(ctx context.Context, args string) (string, error)
}

// StartRunSpan starts an action run span with common trace attributes.
func StartRunSpan(ctx context.Context, tracerName, spanName, args string) (context.Context, trace.Span) {
	return otel.Tracer(tracerName).Start(ctx, spanName, trace.WithAttributes(
		attribute.String("action.args", args),
	))
}
