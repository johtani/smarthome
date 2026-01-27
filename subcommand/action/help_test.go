package action

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"testing"
)

func TestHelpAction_Run(t *testing.T) {
	// TracerProviderのモック
	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithSyncer(exporter),
	)
	otel.SetTracerProvider(tp)

	msg := "help message"
	a := NewHelpAction(msg)
	got, err := a.Run(context.Background(), "")
	if err != nil {
		t.Errorf("Run() error = %v, wantErr %v", err, false)
		return
	}
	if got != msg {
		t.Errorf("Run() got = %v, want %v", got, msg)
	}

	// Spanの確認
	spans := exporter.GetSpans()
	found := false
	for _, s := range spans {
		if s.Name == "HelpAction.Run" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'HelpAction.Run' span to be recorded, but not found")
	}
}
