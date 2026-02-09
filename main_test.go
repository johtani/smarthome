package main

import (
	"context"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/johtani/smarthome/subcommand"
	"github.com/johtani/smarthome/subcommand/action/owntone"
	"github.com/johtani/smarthome/subcommand/action/switchbot"
	"github.com/johtani/smarthome/subcommand/action/yamaha"
)

func TestRunCmd(t *testing.T) {
	config := subcommand.Config{
		Owntone:   owntone.Config{Url: "http://localhost:8000"},
		Switchbot: switchbot.Config{Token: "token", Secret: "secret"},
		Yamaha:    yamaha.Config{Url: "http://localhost:8080"},
		Commands:  subcommand.NewCommands(),
	}

	ctx := context.Background()

	tests := []struct {
		name     string
		args     []string
		wantCont string
		wantErr  bool
	}{
		{
			name:     "too few args",
			args:     []string{"smarthome"},
			wantCont: "コマンドを指定してください",
			wantErr:  true,
		},
		{
			name:     "help command",
			args:     []string{"smarthome", "help"},
			wantCont: "利用可能なコマンドは次の通りです",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout and stderr
			oldStdout := os.Stdout
			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stdout = w
			os.Stderr = w

			err := runCmd(ctx, config, tt.args)

			w.Close()
			out, _ := io.ReadAll(r)
			os.Stdout = oldStdout
			os.Stderr = oldStderr

			if (err != nil) != tt.wantErr {
				t.Errorf("runCmd() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !strings.Contains(string(out), tt.wantCont) && (err != nil && !strings.Contains(err.Error(), tt.wantCont)) {
				t.Errorf("runCmd() output/error = %q, want context %q", string(out), tt.wantCont)
			}
		})
	}
}
