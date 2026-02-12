package switchbot

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/nasa9084/go-switchbot/v3"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"sync"
)

type Config struct {
	Token string `json:"token"`
	// #nosec G117
	Secret           string `json:"secret"`
	LightDeviceID    string `json:"light_device_id"`
	LightSceneID     string `json:"light_scene_id"`
	AirConditionerID string `json:"air_conditioner_id"`
}

func (c Config) Validate() error {
	var errs []string
	if len(c.Token) == 0 {
		errs = append(errs, "switchbot.token is required")
	}
	if len(c.Secret) == 0 {
		errs = append(errs, "switchbot.secret is required")
	}
	if len(errs) > 0 {
		return fmt.Errorf("switchbot config validation failed: %s", strings.Join(errs, ", "))
	}
	return nil
}

type DeviceAPI interface {
	List(ctx context.Context) ([]switchbot.Device, []switchbot.InfraredDevice, error)
	Status(ctx context.Context, id string) (switchbot.DeviceStatus, error)
	Command(ctx context.Context, id string, cmd switchbot.Command) error
}

type SceneAPI interface {
	List(ctx context.Context) ([]switchbot.Scene, error)
	Execute(ctx context.Context, id string) error
}

type CachedClient struct {
	DeviceAPI
	SceneAPI
	deviceNameCache map[string]string
	sceneNameCache  map[string]string
	mu              sync.RWMutex
}

func NewClient(config Config) *CachedClient {
	c := switchbot.New(config.Token, config.Secret, switchbot.WithHTTPClient(&http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}))
	return &CachedClient{
		DeviceAPI:       c.Device(),
		SceneAPI:        c.Scene(),
		deviceNameCache: map[string]string{},
		sceneNameCache:  map[string]string{},
	}
}

func (c *CachedClient) GetSceneName(ctx context.Context, id string) (string, error) {
	c.mu.RLock()
	name, ok := c.sceneNameCache[id]
	c.mu.RUnlock()
	if ok {
		return name, nil
	}
	scenes, err := c.SceneAPI.List(ctx)
	if err != nil {
		return "", err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, scene := range scenes {
		c.sceneNameCache[scene.ID] = scene.Name
	}
	return c.sceneNameCache[id], nil
}

func (c *CachedClient) GetDeviceName(ctx context.Context, id string) (string, error) {
	c.mu.RLock()
	name, ok := c.deviceNameCache[id]
	c.mu.RUnlock()
	if ok {
		return name, nil
	}
	pDevices, vDevices, err := c.DeviceAPI.List(ctx)
	if err != nil {
		return "", err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
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
