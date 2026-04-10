package subcommand

import (
	"context"
	"strings"
	"testing"
)

func TestNewHelpSubcommand(t *testing.T) {
	config := Config{
		Commands: Commands{
			Definitions: []Definition{
				{
					Name:        "sample",
					Description: "sample description",
					Factory:     NewDummySubcommand,
				},
			},
		},
	}

	sub := NewHelpSubcommand(NewHelpDefinition(), config)

	got, err := sub.Exec(context.Background(), "")
	if err != nil {
		t.Fatalf("Exec() error = %v", err)
	}

	if !strings.HasPrefix(got, "利用可能なコマンドは次の通りです\n") {
		t.Fatalf("help should start with guidance line, got: %q", got)
	}
	if !strings.Contains(got, "  sample : sample description\n") {
		t.Fatalf("help should include command list, got: %q", got)
	}
	if !strings.Contains(got, "commit: ") {
		t.Fatalf("help should include commit/version info, got: %q", got)
	}
}
