package subcommand

import (
	"runtime/debug"
	"testing"
)

func TestRevisionFromBuildSettings(t *testing.T) {
	tests := []struct {
		name     string
		settings []debug.BuildSetting
		want     string
	}{
		{
			name: "no revision falls back to unknown",
			settings: []debug.BuildSetting{
				{Key: "vcs.time", Value: "2026-04-10T00:00:00Z"},
			},
			want: unknownRevision,
		},
		{
			name: "short revision stays as is",
			settings: []debug.BuildSetting{
				{Key: "vcs.revision", Value: "abc1234"},
			},
			want: "abc1234",
		},
		{
			name: "long revision is shortened",
			settings: []debug.BuildSetting{
				{Key: "vcs.revision", Value: "0123456789abcdef"},
			},
			want: "0123456789ab",
		},
		{
			name: "dirty suffix is appended",
			settings: []debug.BuildSetting{
				{Key: "vcs.revision", Value: "0123456789abcdef"},
				{Key: "vcs.modified", Value: "true"},
			},
			want: "0123456789ab-dirty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := revisionFromBuildSettings(tt.settings); got != tt.want {
				t.Errorf("revisionFromBuildSettings() = %q, want %q", got, tt.want)
			}
		})
	}
}
