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

func is_target_device(deviceTypes []string, target string) bool {
	for _, item := range deviceTypes {
		if target == item {
			return true
		}
	}
	return false
}

func (a GetTemperatureAndHumidityAction) Run() (string, error) {
	msg := map[string]string{}
	pdev, vdev, err := a.Device().List(context.Background())
	if err != nil {
		return "", err
	}
	for _, d := range pdev {
		if is_target_device(a.deviceTypes, string(d.Type)) {
			status, err := a.Device().Status(context.Background(), d.ID)
			if err != nil {
				return "", err
			}
			msg[d.Name] = fmt.Sprintf("%s(🔋%d)\t%.1f℃ \t%d％", d.Name, status.Battery, status.Temperature, status.Humidity)
		}
	}
	for _, d := range vdev {
		if is_target_device(a.deviceTypes, string(d.Type)) {
			status, err := a.Device().Status(context.Background(), d.ID)
			if err != nil {
				return "", err
			}
			msg[d.Name] = fmt.Sprintf("%s(🔋%d)\t%.1f℃ \t%d％", d.Name, status.Battery, status.Temperature, status.Humidity)
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
		"Get temperature and humidity from devices on SwitchBot",
		[]string{"Meter", "WoIOSensor"},
		client,
	}
}
