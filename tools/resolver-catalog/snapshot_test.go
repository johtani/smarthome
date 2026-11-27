package main

import (
	"testing"

	"github.com/johtani/smarthome/subcommand"
)

func TestSnapshotPreservesRuntimeArgumentHints(t *testing.T) {
	for _, d := range subcommand.NewCommands().Definitions {
		for _, explicit := range []bool{false, true} {
			defs, err := snapshotDefinitions([]entry{{Name: d.Name, Args: d.ArgHints(explicit)}})
			if err != nil {
				t.Fatal(err)
			}
			if got := defs[d.Name].ArgHints(true); got != d.ArgHints(true) {
				t.Errorf("%s: restored hints %q, want %q", d.Name, got, d.ArgHints(true))
			}
		}
	}
}

func TestSnapshotRejectsUnknownHintFormat(t *testing.T) {
	if _, err := snapshotDefinitions([]entry{{Name: "example", Args: "<query>"}}); err == nil {
		t.Fatal("expected invalid argument hint error")
	}
}
