package cron

import (
	"context"
	"fmt"
	"github.com/johtani/smarthome/server/cron/influxdb"
	"github.com/johtani/smarthome/subcommand/action/switchbot"
)

func RecordTemp(influxdbConfig influxdb.Config, switchbotConfig switchbot.Config) {
	fmt.Println("Run RecordTemp")
	targetTypes := []string{"Meter", "WoIOSensor", "MeterPlus"}
	sCli := switchbot.NewClient(switchbotConfig)
	iCli := influxdb.NewClient(influxdbConfig)

	pdev, vdev, err := sCli.Device().List(context.Background())
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, d := range pdev {
		if switchbot.IsTargetDevice(targetTypes, string(d.Type)) {
			status, err := sCli.Device().Status(context.Background(), d.ID)
			if err != nil {
				fmt.Println(err)
			}
			data := influxdb.Temperature{
				Room:        d.Name,
				Temperature: status.Temperature,
				Humidity:    status.Humidity,
				Battery:     status.Battery,
				Co2:         -1,
			}
			iCli.WriteTemperature(data)
		}
	}
	for _, d := range vdev {
		if switchbot.IsTargetDevice(targetTypes, string(d.Type)) {
			status, err := sCli.Device().Status(context.Background(), d.ID)
			if err != nil {
				fmt.Println(err)
			}

			data := influxdb.Temperature{
				Room:        d.Name,
				Temperature: status.Temperature,
				Humidity:    status.Humidity,
				Battery:     status.Battery,
				Co2:         -1,
			}
			iCli.WriteTemperature(data)
		}
	}
}
