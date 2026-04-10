package subcommand

import "runtime/debug"

const (
	unknownRevision  = "unknown"
	revisionShortLen = 12
)

// currentRevision returns the current VCS revision from Go build info.
// When unavailable, it falls back to "unknown".
func currentRevision() string {
	bi, ok := debug.ReadBuildInfo()
	if !ok || bi == nil {
		return unknownRevision
	}

	return revisionFromBuildSettings(bi.Settings)
}

func revisionFromBuildSettings(settings []debug.BuildSetting) string {
	revision := ""
	modified := false
	for _, s := range settings {
		switch s.Key {
		case "vcs.revision":
			revision = s.Value
		case "vcs.modified":
			modified = s.Value == "true"
		}
	}

	if revision == "" {
		return unknownRevision
	}
	if len(revision) > revisionShortLen {
		revision = revision[:revisionShortLen]
	}
	if modified {
		return revision + "-dirty"
	}

	return revision
}
