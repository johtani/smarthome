/*
Package cron provides scheduled tasks for the smart home system.
It currently handles periodic recording of temperature and humidity.
*/
package cron

import (
	"fmt"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/johtani/smarthome/internal/configstore"
)

// Start registers the temperature job and starts its scheduler.
// The caller must shut down the returned scheduler.
func Start(configStore *configstore.Store) (gocron.Scheduler, error) {
	return start(configStore, "Asia/Tokyo", "*/10 * * * *", gocron.NewScheduler)
}

func start(configStore *configstore.Store, zone, schedule string, newScheduler func(...gocron.SchedulerOption) (gocron.Scheduler, error)) (gocron.Scheduler, error) {
	jst, err := time.LoadLocation(zone)
	if err != nil {
		return nil, fmt.Errorf("タイムゾーンの読み込みに失敗しました: %w", err)
	}
	s, err := newScheduler(gocron.WithLocation(jst))
	if err != nil {
		return nil, fmt.Errorf("スケジューラーの初期化に失敗しました: %w", err)
	}
	_, err = s.NewJob(
		gocron.CronJob(schedule, false),
		gocron.NewTask(
			RecordTemp,
			configStore,
		),
	)
	if err != nil {
		_ = s.Shutdown()
		return nil, fmt.Errorf("ジョブの登録に失敗しました: %w", err)
	}
	s.Start()
	fmt.Println("Start cron service...")
	return s, nil
}
