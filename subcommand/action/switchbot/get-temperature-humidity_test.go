package switchbot

import (
	"context"
	"strings"
	"testing"

	"github.com/nasa9084/go-switchbot/v3"
)

type mockDeviceAPI struct {
	listFunc   func(ctx context.Context) ([]switchbot.Device, []switchbot.InfraredDevice, error)
	statusFunc func(ctx context.Context, id string) (switchbot.DeviceStatus, error)
}

func (m *mockDeviceAPI) List(ctx context.Context) ([]switchbot.Device, []switchbot.InfraredDevice, error) {
	return m.listFunc(ctx)
}
func (m *mockDeviceAPI) Status(ctx context.Context, id string) (switchbot.DeviceStatus, error) {
	return m.statusFunc(ctx, id)
}
func (m *mockDeviceAPI) Command(_ context.Context, _ string, _ switchbot.Command) error {
	return nil
}

func TestGetTemperatureAndHumidityAction_Run(t *testing.T) {
	mock := &mockDeviceAPI{
		listFunc: func(_ context.Context) ([]switchbot.Device, []switchbot.InfraredDevice, error) {
			return []switchbot.Device{
				{ID: "1", Name: "Meter 1", Type: "Meter"},
				{ID: "2", Name: "CO2 Meter", Type: "MeterPro(CO2)"},
				{ID: "3", Name: "Light", Type: "Light"}, // Should be filtered out
				{ID: "4", Name: "Weather Station", Type: "WeatherStation"},
			}, nil, nil
		},
		statusFunc: func(_ context.Context, id string) (switchbot.DeviceStatus, error) {
			if id == "1" {
				return switchbot.DeviceStatus{Temperature: 25.5, Humidity: 50, Battery: 90}, nil
			}
			if id == "2" {
				return switchbot.DeviceStatus{Temperature: 20.0, Humidity: 45, Battery: 80, CO2: 800}, nil
			}
			if id == "4" {
				return switchbot.DeviceStatus{Temperature: 22.5, Humidity: 55, Battery: 70}, nil
			}
			return switchbot.DeviceStatus{}, nil
		},
	}

	client := &CachedClient{
		DeviceAPI: mock,
	}
	action := NewGetTemperatureAndHumidityAction(client)

	got, err := action.Run(context.Background(), "")
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	lines := strings.Split(got, "\n")
	if len(lines) != 3 {
		t.Fatalf("Run() returned %d lines, want 3", len(lines))
	}

	// Check sorting and content
	// "CO2 Meter" comes first because of sort.Strings
	if !strings.Contains(lines[0], "CO2: 800ppm") || !strings.Contains(lines[0], "CO2 Meter") {
		t.Errorf("First line unexpected: %s", lines[0])
	}
	if !strings.Contains(lines[1], "25.5℃") || !strings.Contains(lines[1], "Meter 1") {
		t.Errorf("Second line unexpected: %s", lines[1])
	}
	if !strings.Contains(lines[2], "22.5℃") || !strings.Contains(lines[2], "55％") || !strings.Contains(lines[2], "Weather Station(🔋70)") {
		t.Errorf("Third line unexpected: %s", lines[2])
	}
	if strings.Contains(got, "Light") {
		t.Errorf("Run() should not contain 'Light' device")
	}
}
