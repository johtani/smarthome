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

func TestRevisionOrVersion(t *testing.T) {
	tests := []struct {
		name string
		info *debug.BuildInfo
		want string
	}{
		{
			name: "uses vcs revision when available",
			info: &debug.BuildInfo{
				Main: debug.Module{Version: "v9.9.9"},
				Settings: []debug.BuildSetting{
					{Key: "vcs.revision", Value: "0123456789abcdef"},
				},
			},
			want: "0123456789ab",
		},
		{
			name: "falls back to main version when revision is unavailable",
			info: &debug.BuildInfo{
				Main: debug.Module{Version: "v1.2.3"},
				Settings: []debug.BuildSetting{
					{Key: "vcs.time", Value: "2026-04-10T00:00:00Z"},
				},
			},
			want: "v1.2.3",
		},
		{
			name: "returns unknown when main version is devel",
			info: &debug.BuildInfo{
				Main: debug.Module{Version: develVersion},
				Settings: []debug.BuildSetting{
					{Key: "vcs.time", Value: "2026-04-10T00:00:00Z"},
				},
			},
			want: unknownRevision,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := revisionOrVersion(tt.info); got != tt.want {
				t.Errorf("revisionOrVersion() = %q, want %q", got, tt.want)
			}
		})
	}
}
