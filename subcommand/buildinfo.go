package subcommand

import (
	"runtime/debug"
	"strings"
)

const (
	unknownRevision  = "unknown"
	revisionShortLen = 12
	develVersion     = "(devel)"
)

// currentRevision returns the current VCS revision from Go build info.
// If unavailable, it falls back to module version, then "unknown".
func currentRevision() string {
	bi, ok := debug.ReadBuildInfo()
	if !ok || bi == nil {
		return unknownRevision
	}

	return revisionOrVersion(bi)
}

func revisionOrVersion(bi *debug.BuildInfo) string {
	revision := revisionFromBuildSettings(bi.Settings)
	if revision != unknownRevision {
		return revision
	}

	return versionFromBuildInfo(bi)
}

func versionFromBuildInfo(bi *debug.BuildInfo) string {
	version := strings.TrimSpace(bi.Main.Version)
	if version == "" || version == develVersion {
		return unknownRevision
	}

	return version
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
