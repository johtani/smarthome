package switchbot

import (
	"context"
	"fmt"
	"github.com/nasa9084/go-switchbot/v2"
	"strings"
)

const EnvToken = "SWITCHBOT_TOKEN"
const EnvSecret = "SWITCHBOT_SECRET"

type Config struct {
	token  string
	secret string
}

func NewConfig(token string, secret string) (Config, error) {
	var errs []string
	if len(token) == 0 {
		errs = append(errs, fmt.Sprintf("not found \"%s\". Please set %s via Environment variable", EnvToken, EnvToken))
	}
	if len(secret) == 0 {
		errs = append(errs, fmt.Sprintf("not found \"%s\". Please set %s via Environment variable", EnvSecret, EnvSecret))
	}
	if len(errs) > 0 {
		return Config{}, fmt.Errorf(strings.Join(errs, "\n"))
	}
	return Config{
		token,
		secret,
	}, nil
}

type CachedClient struct {
	*switchbot.Client
	deviceNameCache map[string]string
	sceneNameCache  map[string]string
}

func NewClient(config Config) CachedClient {
	return CachedClient{
		switchbot.New(config.token, config.secret),
		map[string]string{},
		map[string]string{},
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
