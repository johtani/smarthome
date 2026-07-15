package subcommand

import (
	"strings"
	"unicode"

	"github.com/johtani/smarthome/subcommand/action/llm"
)

var randomPlaybackCues = []string{
	"ランダム",
	"シャッフル",
	"適当に",
	"おまかせ",
	"お任せ",
	"何か",
	"random",
	"shuffle",
}

var genericPlaybackRequests = map[string]struct{}{
	"音楽をかけて":    {},
	"音楽かけて":     {},
	"音楽を流して":    {},
	"音楽流して":     {},
	"音楽を再生して":   {},
	"音楽再生して":    {},
	"曲をかけて":     {},
	"曲かけて":      {},
	"曲を流して":     {},
	"曲流して":      {},
	"曲を再生して":    {},
	"曲再生して":     {},
	"再生して":      {},
	"playmusic": {},
}

const specifiedMusicTargetCorrectionReason = "specified_music_target"

type resolutionCorrection struct {
	initialCommand   string
	initialArgsKind  string
	commandCorrected bool
	reason           string
}

// normalizeMusicResolution prevents a named music request from being treated
// as random playback when a natural-language resolver selects start music.
func normalizeMusicResolution(input string, resolved llm.ResolvedCommand) llm.ResolvedCommand {
	if resolved.Command != StartMusicCmd || isGenericOrRandomPlaybackRequest(input) {
		return resolved
	}

	resolved.Command = SearchAndPlayMusicCmd
	args := strings.TrimSpace(resolved.Args)
	if args == "" || args == "artist" || args == "genre" {
		resolved.Args = strings.TrimSpace(input)
	} else {
		resolved.Args = args
	}
	return resolved
}

func describeResolutionCorrection(before, after llm.ResolvedCommand) resolutionCorrection {
	correction := resolutionCorrection{
		initialCommand:  before.Command,
		initialArgsKind: classifyResolverArgs(before.Args),
	}
	if before.Command != after.Command {
		correction.commandCorrected = true
		correction.reason = specifiedMusicTargetCorrectionReason
	}
	return correction
}

func classifyResolverArgs(args string) string {
	switch strings.TrimSpace(args) {
	case "":
		return "empty"
	case "artist", "genre":
		return "mode"
	default:
		return "free_text"
	}
}

func isGenericOrRandomPlaybackRequest(input string) bool {
	normalized := normalizePlaybackRequest(input)
	for _, cue := range randomPlaybackCues {
		if strings.Contains(normalized, cue) {
			return true
		}
	}

	for {
		trimmed := strings.TrimSuffix(normalized, "ください")
		trimmed = strings.TrimSuffix(trimmed, "下さい")
		trimmed = strings.TrimSuffix(trimmed, "お願い")
		trimmed = strings.TrimSuffix(trimmed, "ね")
		if trimmed == normalized {
			break
		}
		normalized = trimmed
	}

	_, ok := genericPlaybackRequests[normalized]
	return ok
}

func normalizePlaybackRequest(input string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) || unicode.IsPunct(r) {
			return -1
		}
		return unicode.ToLower(r)
	}, strings.TrimSpace(input))
}
