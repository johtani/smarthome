package switchbot

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

type GetTemperatureAndHumidityAction struct {
	name        string
	deviceTypes []string
	CachedClient
}

func isTargetDevice(deviceTypes []string, target string) bool {
	for _, item := range deviceTypes {
		if target == item {
			return true
		}
	}
	return false
}

func (a GetTemperatureAndHumidityAction) Run(_ string) (string, error) {
	msg := map[string]string{}
	//goland:noinspection SpellCheckingInspection
	pdev, vdev, err := a.Device().List(context.Background())
	if err != nil {
		return "", err
	}
	for _, d := range pdev {
		if isTargetDevice(a.deviceTypes, string(d.Type)) {
			status, err := a.Device().Status(context.Background(), d.ID)
			if err != nil {
				return "", err
			}
			msg[d.Name] = fmt.Sprintf("%.1f℃ \t%d％ / %s(🔋%d)", status.Temperature, status.Humidity, d.Name, status.Battery)
		}
	}
	for _, d := range vdev {
		if isTargetDevice(a.deviceTypes, string(d.Type)) {
			status, err := a.Device().Status(context.Background(), d.ID)
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

func NewGetTemperatureAndHumidityAction(client CachedClient) GetTemperatureAndHumidityAction {
	return GetTemperatureAndHumidityAction{
		name:         "Get temperature and humidity from devices on SwitchBot",
		deviceTypes:  []string{"Meter", "WoIOSensor"},
		CachedClient: client,
	}
}
