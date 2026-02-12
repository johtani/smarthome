/*
Package cron provides scheduled tasks for the smart home system.
It currently handles periodic recording of temperature and humidity.
*/
package cron

import (
	"fmt"
	"github.com/go-co-op/gocron/v2"
	"github.com/johtani/smarthome/subcommand"
	"time"
)

// Run starts the cron scheduler and runs scheduled jobs.
func Run(config subcommand.Config) error {
	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return fmt.Errorf("タイムゾーンの読み込みに失敗しました: %w", err)
	}
	s, err := gocron.NewScheduler(gocron.WithLocation(jst))
	if err != nil {
		return fmt.Errorf("スケジューラーの初期化に失敗しました: %w", err)
	}
	_, err = s.NewJob(
		gocron.CronJob("*/10 * * * *", false),
		gocron.NewTask(
			RecordTemp,
			config.Influxdb,
			config.Switchbot,
		),
	)
	if err != nil {
		return fmt.Errorf("ジョブの登録に失敗しました: %w", err)
	}
	s.Start()
	fmt.Println("Start cron service...")
	select {}
}
