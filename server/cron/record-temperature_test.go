package cron

import (
	"context"
	"testing"

	"github.com/johtani/smarthome/server/cron/influxdb"
	sb "github.com/johtani/smarthome/subcommand/action/switchbot"
	"github.com/nasa9084/go-switchbot/v3"
)

type mockDeviceAPI struct {
	sb.DeviceAPI
	list   func(ctx context.Context) ([]switchbot.Device, []switchbot.InfraredDevice, error)
	status func(ctx context.Context, id string) (switchbot.DeviceStatus, error)
}

func (m *mockDeviceAPI) List(ctx context.Context) ([]switchbot.Device, []switchbot.InfraredDevice, error) {
	return m.list(ctx)
}

func (m *mockDeviceAPI) Status(ctx context.Context, id string) (switchbot.DeviceStatus, error) {
	return m.status(ctx, id)
}

type mockInfluxDBClient struct {
	influxdb.Client
	writeTemperature func(data influxdb.Temperature)
}

func (m *mockInfluxDBClient) WriteTemperature(data influxdb.Temperature) {
	m.writeTemperature(data)
}

func (m *mockInfluxDBClient) Close() {}

func TestExecuteRecordTemp(t *testing.T) {
	ctx := context.Background()

	sCli := &sb.CachedClient{
		DeviceAPI: &mockDeviceAPI{
			list: func(ctx context.Context) ([]switchbot.Device, []switchbot.InfraredDevice, error) {
				return []switchbot.Device{
					{
						ID:   "meter-id",
						Name: "Living Room",
						Type: switchbot.Meter,
					},
					{
						ID:   "plug-id",
						Name: "Plug",
						Type: switchbot.Plug,
					},
				}, nil, nil
			},
			status: func(ctx context.Context, id string) (switchbot.DeviceStatus, error) {
				if id == "meter-id" {
					return switchbot.DeviceStatus{
						Temperature: 25.5,
						Humidity:    50,
						Battery:     90,
					}, nil
				}
				return switchbot.DeviceStatus{}, nil
			},
		},
	}

	written := false
	iCli := &mockInfluxDBClient{
		writeTemperature: func(data influxdb.Temperature) {
			written = true
			if data.Room != "Living Room" {
				t.Errorf("expected Living Room, got %s", data.Room)
			}
			if data.Temperature != 25.5 {
				t.Errorf("expected 25.5, got %f", data.Temperature)
			}
		},
	}

	ExecuteRecordTemp(ctx, sCli, iCli)

	if !written {
		t.Error("expected WriteTemperature to be called")
	}
}
