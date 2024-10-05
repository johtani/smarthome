package cron

import (
	"github.com/go-co-op/gocron/v2"
	"github.com/johtani/smarthome/subcommand"
	"time"
)

func Run(config subcommand.Config) {
	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		panic(err)
	}
	s, err := gocron.NewScheduler(gocron.WithLocation(jst))
	if err != nil {
		panic(err)
	}
	_, err = s.NewJob(
		gocron.CronJob("0,10 * * * *", false),
		gocron.NewTask(
			RecordTemp,
			config.Influxdb,
			config.Switchbot,
		),
	)
	if err != nil {
		panic(err)
	}
	s.Start()
	select {}
}
