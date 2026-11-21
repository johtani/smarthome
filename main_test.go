package main

import (
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/johtani/smarthome/internal/configstore"
	"github.com/johtani/smarthome/server/cron"
	"github.com/johtani/smarthome/subcommand"
	"github.com/johtani/smarthome/subcommand/action/owntone"
	"github.com/johtani/smarthome/subcommand/action/switchbot"
	"github.com/johtani/smarthome/subcommand/action/yamaha"
)

func TestRunServerDelayedCronFailure(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		release := make(chan struct{})
		result := make(chan error, 1)
		cause := errors.New("cron initialization failed")
		slackCalled := false
		go func() {
			result <- runServer(nil, func(*configstore.Store) (gocron.Scheduler, error) {
				<-release
				return nil, cause
			}, func(*configstore.Store) error {
				slackCalled = true
				return nil
			})
		}()
		synctest.Wait()
		close(release)
		if err := <-result; !errors.Is(err, cause) {
			t.Errorf("runServer returned %v, want wrapped %v", err, cause)
		}
		if slackCalled {
			t.Error("Slack started despite cron initialization failure")
		}
	})
}

func TestRunServerStartsTemperatureJobBeforeSlack(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var scheduler gocron.Scheduler
		startedAt := time.Now()
		slackResult := errors.New("Slack returned")
		err := runServer(nil, func(store *configstore.Store) (gocron.Scheduler, error) {
			var err error
			scheduler, err = cron.Start(store)
			return scheduler, err
		}, func(*configstore.Store) error {
			jobs := scheduler.Jobs()
			if len(jobs) != 1 {
				t.Fatalf("got %d registered jobs, want 1", len(jobs))
			}
			runs, err := jobs[0].NextRuns(2)
			if err != nil || len(runs) != 2 {
				t.Fatalf("NextRuns(2) = %v, %v", runs, err)
			}
			if runs[0].Location().String() != "Asia/Tokyo" || runs[0].Minute()%10 != 0 || runs[0].Second() != 0 || runs[1].Sub(runs[0]) != 10*time.Minute {
				t.Errorf("unexpected temperature job schedule: %v", runs)
			}
			if !time.Now().Equal(startedAt) {
				t.Error("startup advanced the test clock before Slack started")
			}
			return slackResult
		})
		if !errors.Is(err, slackResult) {
			t.Errorf("runServer returned %v, want %v", err, slackResult)
		}
	})
}

func TestRunCmd(t *testing.T) {
	config := subcommand.Config{
		Owntone:   owntone.Config{URL: "http://localhost:8000"},
		Switchbot: switchbot.Config{Token: "token", Secret: "secret"},
		Yamaha:    yamaha.Config{URL: "http://localhost:8080"},
		Commands:  subcommand.NewCommands(),
	}
	store := configstore.New(config)

	ctx := context.Background()

	tests := []struct {
		name     string
		args     []string
		wantCont string
		wantErr  bool
	}{
		{
			name:     "no args",
			args:     []string{},
			wantCont: "コマンドを指定してください",
			wantErr:  true,
		},
		{
			name:     "help command",
			args:     []string{"help"},
			wantCont: "commit:",
			wantErr:  false,
		},
		{
			name:     "multi word command is accepted by arg guard",
			args:     []string{"unknown", "command"},
			wantCont: "sorry, i cannot understand",
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

			err := runCmd(ctx, store, tt.args)
			_ = w.Close()
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
