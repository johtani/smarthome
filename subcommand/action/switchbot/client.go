package switchbot

import (
	"context"
	"fmt"
	"github.com/nasa9084/go-switchbot/v3"
	"strings"
)

type Config struct {
	Token            string `json:"token"`
	Secret           string `json:"secret"`
	LightDeviceId    string `json:"light_device_id"`
	LightSceneId     string `json:"light_scene_id"`
	AirConditionerId string `json:"air_conditioner_id"`
}

func (c Config) Validate() error {
	var errs []string
	if len(c.Token) == 0 {
		errs = append(errs, fmt.Sprintf("not found \"switchbot.Token\". Please check config file."))
	}
	if len(c.Secret) == 0 {
		errs = append(errs, fmt.Sprintf("not found \"switchbot.Secret\". Please check config file."))
	}
	if len(errs) > 0 {
		return fmt.Errorf(strings.Join(errs, "\n"))
	}
	return nil
}

type CachedClient struct {
	*switchbot.Client
	deviceNameCache map[string]string
	sceneNameCache  map[string]string
}

func NewClient(config Config) CachedClient {
	return CachedClient{
		Client:          switchbot.New(config.Token, config.Secret),
		deviceNameCache: map[string]string{},
		sceneNameCache:  map[string]string{},
	}
}

func (c CachedClient) GetSceneName(id string) (string, error) {
	name, ok := c.sceneNameCache[id]
	if ok {
		return name, nil
	}
	scenes, err := c.Scene().List(context.Background())
	if err != nil {
		return "", err
	}
	c.sceneNameCache = map[string]string{}
	for _, scene := range scenes {
		c.sceneNameCache[scene.ID] = scene.Name
	}
	return c.sceneNameCache[id], nil
}

func (c CachedClient) GetDeviceName(id string) (string, error) {
	name, ok := c.deviceNameCache[id]
	if ok {
		return name, nil
	}
	pDevices, vDevices, err := c.Device().List(context.Background())
	if err != nil {
		return "", err
	}
	c.deviceNameCache = map[string]string{}
	for _, device := range pDevices {
		c.deviceNameCache[device.ID] = device.Name
	}
	for _, device := range vDevices {
		c.deviceNameCache[device.ID] = device.Name
	}
	return c.deviceNameCache[id], nil
}

func IsTargetDevice(deviceTypes []string, target string) bool {
	for _, item := range deviceTypes {
		if target == item {
			return true
		}
	}
	return false
}
