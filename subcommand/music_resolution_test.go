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
