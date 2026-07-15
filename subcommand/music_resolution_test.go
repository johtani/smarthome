package subcommand

import (
	"testing"

	"github.com/johtani/smarthome/subcommand/action/llm"
)

func TestNormalizeMusicResolution(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		resolved    llm.ResolvedCommand
		wantCommand string
		wantArgs    string
	}{
		{
			name:        "keeps free text search argument",
			input:       "B'zの曲をかけて",
			resolved:    llm.ResolvedCommand{Command: StartMusicCmd, Args: "B'z"},
			wantCommand: SearchAndPlayMusicCmd,
			wantArgs:    "B'z",
		},
		{
			name:        "restores input for artist enum",
			input:       "B'zの曲をかけて",
			resolved:    llm.ResolvedCommand{Command: StartMusicCmd, Args: "artist"},
			wantCommand: SearchAndPlayMusicCmd,
			wantArgs:    "B'zの曲をかけて",
		},
		{
			name:        "restores input for genre enum",
			input:       "ジャズをかけて",
			resolved:    llm.ResolvedCommand{Command: StartMusicCmd, Args: "genre"},
			wantCommand: SearchAndPlayMusicCmd,
			wantArgs:    "ジャズをかけて",
		},
		{
			name:        "restores input for empty args",
			input:       "First Loveを再生して",
			resolved:    llm.ResolvedCommand{Command: StartMusicCmd},
			wantCommand: SearchAndPlayMusicCmd,
			wantArgs:    "First Loveを再生して",
		},
		{
			name:        "keeps generic music playback",
			input:       "音楽をかけてください",
			resolved:    llm.ResolvedCommand{Command: StartMusicCmd},
			wantCommand: StartMusicCmd,
			wantArgs:    "",
		},
		{
			name:        "keeps random playback",
			input:       "適当に何か流して",
			resolved:    llm.ResolvedCommand{Command: StartMusicCmd},
			wantCommand: StartMusicCmd,
			wantArgs:    "",
		},
		{
			name:        "keeps random genre playback",
			input:       "ランダムにジャズを流して",
			resolved:    llm.ResolvedCommand{Command: StartMusicCmd, Args: "genre"},
			wantCommand: StartMusicCmd,
			wantArgs:    "genre",
		},
		{
			name:        "does not change other commands",
			input:       "B'zの曲をかけて",
			resolved:    llm.ResolvedCommand{Command: SearchAndPlayMusicCmd, Args: "B'z"},
			wantCommand: SearchAndPlayMusicCmd,
			wantArgs:    "B'z",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeMusicResolution(tt.input, tt.resolved)
			if got.Command != tt.wantCommand {
				t.Fatalf("command = %q, want %q", got.Command, tt.wantCommand)
			}
			if got.Args != tt.wantArgs {
				t.Fatalf("args = %q, want %q", got.Args, tt.wantArgs)
			}
		})
	}
}

func TestDescribeResolutionCorrection(t *testing.T) {
	tests := []struct {
		name          string
		before        llm.ResolvedCommand
		after         llm.ResolvedCommand
		wantArgsKind  string
		wantCorrected bool
		wantReason    string
	}{
		{
			name:          "mode args corrected",
			before:        llm.ResolvedCommand{Command: StartMusicCmd, Args: "artist"},
			after:         llm.ResolvedCommand{Command: SearchAndPlayMusicCmd, Args: "B'zの曲をかけて"},
			wantArgsKind:  "mode",
			wantCorrected: true,
			wantReason:    specifiedMusicTargetCorrectionReason,
		},
		{
			name:         "empty args unchanged",
			before:       llm.ResolvedCommand{Command: StartMusicCmd},
			after:        llm.ResolvedCommand{Command: StartMusicCmd},
			wantArgsKind: "empty",
		},
		{
			name:         "free text args unchanged",
			before:       llm.ResolvedCommand{Command: SearchAndPlayMusicCmd, Args: "B'z"},
			after:        llm.ResolvedCommand{Command: SearchAndPlayMusicCmd, Args: "B'z"},
			wantArgsKind: "free_text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := describeResolutionCorrection(tt.before, tt.after)
			if got.initialCommand != tt.before.Command {
				t.Fatalf("initial command = %q, want %q", got.initialCommand, tt.before.Command)
			}
			if got.initialArgsKind != tt.wantArgsKind {
				t.Fatalf("initial args kind = %q, want %q", got.initialArgsKind, tt.wantArgsKind)
			}
			if got.commandCorrected != tt.wantCorrected {
				t.Fatalf("command corrected = %t, want %t", got.commandCorrected, tt.wantCorrected)
			}
			if got.reason != tt.wantReason {
				t.Fatalf("reason = %q, want %q", got.reason, tt.wantReason)
			}
		})
	}
}
