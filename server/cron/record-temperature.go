package cron

import (
	"context"
	"log/slog"

	"github.com/johtani/smarthome/server/cron/influxdb"
	"github.com/johtani/smarthome/subcommand/action/switchbot"
	"go.opentelemetry.io/otel"
)

func RecordTemp(influxdbConfig influxdb.Config, switchbotConfig switchbot.Config) {
	sCli := switchbot.NewClient(switchbotConfig)
	iCli := influxdb.NewClient(influxdbConfig)
	defer iCli.Close()

	ExecuteRecordTemp(context.Background(), sCli, iCli)
}

func ExecuteRecordTemp(ctx context.Context, sCli *switchbot.CachedClient, iCli influxdb.Client) {
	ctx, span := otel.Tracer("cron").Start(ctx, "RecordTemp")
	defer span.End()

	targetTypes := []string{"Meter", "WoIOSensor", "MeterPlus", "MeterPro(CO2)"}

	pdev, vdev, err := sCli.DeviceAPI.List(ctx)
	if err != nil {
		slog.Error("Cannot get device list", "error", err)
		return
	}
	for _, d := range pdev {
		if switchbot.IsTargetDevice(targetTypes, string(d.Type)) {
			status, err := sCli.DeviceAPI.Status(ctx, d.ID)
			if err != nil {
				slog.Error("Something wrong on getting status", "device", d.Name, "error", err)
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
				slog.Error("Something wrong on getting status", "device", d.Name, "error", err)
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
}
