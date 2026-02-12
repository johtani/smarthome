package cron

import (
	"context"
	"log/slog"

	"github.com/johtani/smarthome/server/cron/influxdb"
	"github.com/johtani/smarthome/subcommand/action/switchbot"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var (
	meter     = otel.Meter("cron")
	tempGauge metric.Float64Gauge
	humGauge  metric.Int64Gauge
	battGauge metric.Int64Gauge
	co2Gauge  metric.Int64Gauge
)

func init() {
	var err error
	tempGauge, err = meter.Float64Gauge("home.temperature", metric.WithDescription("Temperature from SwitchBot Meter"), metric.WithUnit("Celsius"))
	if err != nil {
		slog.Error("Failed to create temperature gauge", "error", err)
	}
	humGauge, err = meter.Int64Gauge("home.humidity", metric.WithDescription("Humidity from SwitchBot Meter"), metric.WithUnit("%"))
	if err != nil {
		slog.Error("Failed to create humidity gauge", "error", err)
	}
	battGauge, err = meter.Int64Gauge("home.battery", metric.WithDescription("Battery level from SwitchBot Meter"), metric.WithUnit("%"))
	if err != nil {
		slog.Error("Failed to create battery gauge", "error", err)
	}
	co2Gauge, err = meter.Int64Gauge("home.co2", metric.WithDescription("CO2 level from SwitchBot Meter"), metric.WithUnit("ppm"))
	if err != nil {
		slog.Error("Failed to create co2 gauge", "error", err)
	}
}

// RecordTemp fetches temperature and humidity from SwitchBot and records it to InfluxDB.
func RecordTemp(influxdbConfig influxdb.Config, switchbotConfig switchbot.Config) {
	sCli := switchbot.NewClient(switchbotConfig)
	iCli := influxdb.NewClient(influxdbConfig)
	defer iCli.Close()

	ExecuteRecordTemp(context.Background(), sCli, iCli)
}

// ExecuteRecordTemp executes the temperature recording logic.
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

			// Record Metrics
			attrs := metric.WithAttributes(attribute.String("room", d.Name))
			tempGauge.Record(ctx, status.Temperature, attrs)
			humGauge.Record(ctx, int64(status.Humidity), attrs)
			battGauge.Record(ctx, int64(status.Battery), attrs)
			if status.CO2 > 0 {
				co2Gauge.Record(ctx, int64(status.CO2), attrs)
			}
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

			// Record Metrics
			attrs := metric.WithAttributes(attribute.String("room", d.Name))
			tempGauge.Record(ctx, status.Temperature, attrs)
			humGauge.Record(ctx, int64(status.Humidity), attrs)
			battGauge.Record(ctx, int64(status.Battery), attrs)
			if status.CO2 > 0 {
				co2Gauge.Record(ctx, int64(status.CO2), attrs)
			}
		}
	}
}
