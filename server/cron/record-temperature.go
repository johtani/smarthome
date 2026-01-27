package cron

import (
	"context"
	"fmt"
	"github.com/johtani/smarthome/server/cron/influxdb"
	"github.com/johtani/smarthome/subcommand/action/switchbot"
	"go.opentelemetry.io/otel"
)

func RecordTemp(influxdbConfig influxdb.Config, switchbotConfig switchbot.Config) {
	ctx, span := otel.Tracer("cron").Start(context.Background(), "RecordTemp")
	defer span.End()

	targetTypes := []string{"Meter", "WoIOSensor", "MeterPlus", "MeterPro(CO2)"}
	sCli := switchbot.NewClient(switchbotConfig)
	iCli := influxdb.NewClient(influxdbConfig)
	defer iCli.Close()

	pdev, vdev, err := sCli.DeviceAPI.List(ctx)
	if err != nil {
		fmt.Printf("Cannot get device list / %v\n", err)
		return
	}
	for _, d := range pdev {
		if switchbot.IsTargetDevice(targetTypes, string(d.Type)) {
			status, err := sCli.DeviceAPI.Status(ctx, d.ID)
			if err != nil {
				fmt.Printf("Something wrong on [%s] / %v\n", d.Name, err)
			}
			data := influxdb.Temperature{
				Room:        d.Name,
				Temperature: status.Temperature,
				Humidity:    status.Humidity,
				Battery:     status.Battery,
				Co2:         status.CO2,
			}
			iCli.WriteTemperature(data)
		}
	}
	for _, d := range vdev {
		if switchbot.IsTargetDevice(targetTypes, string(d.Type)) {
			status, err := sCli.DeviceAPI.Status(ctx, d.ID)
			if err != nil {
				fmt.Printf("Something wrong on [%s] / %v\n", d.Name, err)
			}

			data := influxdb.Temperature{
				Room:        d.Name,
				Temperature: status.Temperature,
				Humidity:    status.Humidity,
				Battery:     status.Battery,
				Co2:         status.CO2,
			}
			iCli.WriteTemperature(data)
		}
	}
	//fmt.Println("Run RecordTemp")
}
