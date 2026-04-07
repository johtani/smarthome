package subcommand

import (
	"context"
	"testing"

	"github.com/johtani/smarthome/subcommand/action"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestSubcommand_Exec_Tracing(t *testing.T) {
	// TracerProviderのモック（メモリ内のExporterを使用）
	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithSyncer(exporter),
	)
	otel.SetTracerProvider(tp)

	s := Subcommand{
		Definition: Definition{
			Name:        "test",
			Description: "test",
		},
		actions: []action.Action{okAction{}},
	}

	ctx := context.Background()
	_, err := s.Exec(ctx, "")
	if err != nil {
		t.Fatalf("Exec() failed: %v", err)
	}

	// Spanが記録されているか確認
	spans := exporter.GetSpans()
	if len(spans) == 0 {
		t.Error("expected spans to be recorded, but got none")
	}

	foundExecSpan := false
	for _, span := range spans {
		if span.Name == "Subcommand.Exec" {
			foundExecSpan = true
			break
		}
	}

	if !foundExecSpan {
		t.Error("expected 'Subcommand.Exec' span to be recorded, but not found")
	}
}

func TestSubcommand_Exec_TracingActionArgsWithFixedArgs(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithSyncer(exporter),
	)
	otel.SetTracerProvider(tp)

	s := Subcommand{
		Definition: Definition{
			Name:        "test",
			Description: "test",
		},
		actions: []action.Action{
			action.NewFixedArgsAction(action.NewHelpAction("ok"), "fixed-arg"),
		},
	}

	if _, err := s.Exec(context.Background(), "original-arg"); err != nil {
		t.Fatalf("Exec() failed: %v", err)
	}

	spans := exporter.GetSpans()
	for _, span := range spans {
		if span.Name != "HelpAction.Run" {
			continue
		}
		for _, attr := range span.Attributes {
			if string(attr.Key) == "action.args" && attr.Value.AsString() == "fixed-arg" {
				return
			}
		}
		t.Fatalf("expected HelpAction.Run span to include action.args=fixed-arg")
	}

	t.Fatalf("expected HelpAction.Run span to be recorded, but not found")
}
