package subcommand

import (
	"testing"

	"github.com/johtani/smarthome/subcommand/action/owntone"
)

func TestBuildSearchAndPlayOptions(t *testing.T) {
	t.Run("without explicit threshold", func(t *testing.T) {
		cfg := owntone.Config{}
		opts := buildSearchAndPlayOptions(cfg, nil)
		if len(opts) != 1 {
			t.Fatalf("expected 1 option, got %d", len(opts))
		}
	})

	t.Run("with explicit threshold", func(t *testing.T) {
		cfg := owntone.Config{
			MusicIntentConfidenceThreshold:    0,
			MusicIntentConfidenceThresholdSet: true,
		}
		opts := buildSearchAndPlayOptions(cfg, nil)
		if len(opts) != 2 {
			t.Fatalf("expected 2 options, got %d", len(opts))
		}
	})
}
