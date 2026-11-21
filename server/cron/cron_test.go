package cron

import (
	"errors"
	"strings"
	"testing"

	"github.com/go-co-op/gocron/v2"
)

func TestStartReportsInitializationErrors(t *testing.T) {
	createErr := errors.New("cannot create scheduler")
	for _, tt := range []struct {
		name, zone, schedule, wantStage string
		newScheduler                    func(...gocron.SchedulerOption) (gocron.Scheduler, error)
		wantCause                       error
	}{
		{"timezone", "Invalid/Zone", "*/10 * * * *", "タイムゾーン", gocron.NewScheduler, nil},
		{"scheduler", "UTC", "*/10 * * * *", "スケジューラー", func(...gocron.SchedulerOption) (gocron.Scheduler, error) {
			return nil, createErr
		}, createErr},
		{"job registration", "UTC", "invalid", "ジョブ", gocron.NewScheduler, nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			scheduler, err := start(nil, tt.zone, tt.schedule, tt.newScheduler)
			if scheduler != nil || err == nil || !strings.Contains(err.Error(), tt.wantStage) || errors.Unwrap(err) == nil {
				t.Fatalf("start returned %v, %v; want wrapped %s error", scheduler, err, tt.wantStage)
			}
			if tt.wantCause != nil && !errors.Is(err, tt.wantCause) {
				t.Errorf("got %v, want cause %v", err, tt.wantCause)
			}
		})
	}
}
