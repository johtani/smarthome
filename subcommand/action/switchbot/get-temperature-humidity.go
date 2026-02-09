package switchbot

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"go.opentelemetry.io/otel"
)

var DefaultDeviceTypes = []string{"Meter", "WoIOSensor", "MeterPlus", "MeterPro(CO2)"}

type GetTemperatureAndHumidityAction struct {
	name        string
	deviceTypes []string
	client      *CachedClient
}

func (a GetTemperatureAndHumidityAction) Run(ctx context.Context, _ string) (string, error) {
	ctx, span := otel.Tracer("switchbot").Start(ctx, "GetTemperatureAndHumidityAction.Run")
	defer span.End()
	msg := map[string]string{}
	//goland:noinspection SpellCheckingInspection
	pdev, vdev, err := a.client.DeviceAPI.List(ctx)
	if err != nil {
		return "", err
	}
	for _, d := range pdev {
		if IsTargetDevice(a.deviceTypes, string(d.Type)) {
			status, err := a.client.Status(ctx, d.ID)
			if err != nil {
				return "", err
			}
			if string(d.Type) == "MeterPro(CO2)" {
				msg[d.Name] = fmt.Sprintf("%.1f℃ \t%d％ / %s(🔋%d) / CO2: %dppm", status.Temperature, status.Humidity, d.Name, status.Battery, status.CO2)
			} else {
				msg[d.Name] = fmt.Sprintf("%.1f℃ \t%d％ / %s(🔋%d)", status.Temperature, status.Humidity, d.Name, status.Battery)
			}
		}
	}
	for _, d := range vdev {
		if IsTargetDevice(a.deviceTypes, string(d.Type)) {
			status, err := a.client.Status(ctx, d.ID)
			if err != nil {
				return "", err
			}
			msg[d.Name] = fmt.Sprintf("%.1f℃ \t%d％ / %s(🔋%d)", status.Temperature, status.Humidity, d.Name, status.Battery)
		}
	}
	// sort by keys
	keys := make([]string, 0, len(msg))
	for k := range msg {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var output []string
	for _, k := range keys {
		output = append(output, msg[k])
	}
	return strings.Join(output, "\n"), nil
}

func NewGetTemperatureAndHumidityAction(client *CachedClient) GetTemperatureAndHumidityAction {
	return GetTemperatureAndHumidityAction{
		name:        "Get temperature and humidity from devices on SwitchBot",
		deviceTypes: DefaultDeviceTypes,
		client:      client,
	}
}
