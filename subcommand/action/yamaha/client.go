package yamaha

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/johtani/smarthome/subcommand/action/internal"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// DefaultTimeout is the default timeout for HTTP requests to Yamaha devices.
const DefaultTimeout = 10 * time.Second
const basePath = "YamahaExtendedControl/v1/main/"

// API is an interface for controlling Yamaha devices.
type API interface {
	SetScene(ctx context.Context, scene int) error
	SetVolume(ctx context.Context, volume int) error
	PowerOn(ctx context.Context) error
	PowerOff(ctx context.Context) error
	SetInput(ctx context.Context, input string) error
}

// Client is a client for the Yamaha MusicCast API.
type Client struct {
	config Config
	http.Client
}

// Config is the configuration for the Yamaha client.
type Config struct {
	URL     string `json:"url"`
	Timeout int    `json:"timeout"`
}

// Validate validates the Yamaha configuration.
func (c Config) Validate() error {
	if c.URL == "" {
		return fmt.Errorf("yamaha.url is required")
	}
	return nil
}

func (c Client) buildURL(path string) string {
	url := c.config.URL
	if !strings.HasSuffix(url, "/") {
		url += "/"
	}
	return url + basePath + path
}

// NewClient creates a new Yamaha client with the given configuration.
func NewClient(config Config) *Client {
	timeout := DefaultTimeout
	if config.Timeout > 0 {
		timeout = time.Duration(config.Timeout) * time.Second
	}
	return &Client{
		config: config,
		Client: http.Client{
			Timeout:   timeout,
			Transport: otelhttp.NewTransport(http.DefaultTransport),
		},
	}
}

// ResponseCode represents the response code from the MusicCast API.
type ResponseCode struct {
	ResponseCode int `json:"response_code"`
}

func parseHTTPResponse(res *http.Response, caller string) error {
	var rc ResponseCode
	if err := internal.DecodeJSONResponse(res, &rc, http.StatusOK); err != nil {
		return err
	}
	if rc.ResponseCode != 0 {
		return fmt.Errorf("something wrong %v... response_code is %v", caller, rc.ResponseCode)
	}
	return nil
}

// SetScene recalls a scene on the Yamaha device.
func (c Client) SetScene(ctx context.Context, scene int) error {
	params := map[string]string{}
	params["num"] = strconv.Itoa(scene)
	req, err := internal.BuildHttpRequestWithParams(ctx, http.MethodGet, c.buildURL("recallScene"), params)
	if err != nil {
		return err
	}
	// #nosec G704
	res, err := c.Do(req)
	if err != nil {
		return err
	}
	err = parseHTTPResponse(res, "SetScene")
	if err != nil {
		return err
	}
	return nil
}

// SetVolume sets the volume level on the Yamaha device.
func (c Client) SetVolume(ctx context.Context, volume int) error {
	params := map[string]string{}
	params["volume"] = strconv.Itoa(volume)
	req, err := internal.BuildHttpRequestWithParams(ctx, http.MethodGet, c.buildURL("setVolume"), params)
	if err != nil {
		return err
	}
	// #nosec G704
	res, err := c.Do(req)
	if err != nil {
		return err
	}
	err = parseHTTPResponse(res, "SetVolume")
	if err != nil {
		return err
	}
	return nil
}

// PowerOn turns on the Yamaha device.
func (c Client) PowerOn(ctx context.Context) error {
	params := map[string]string{}
	params["power"] = "on"
	req, err := internal.BuildHttpRequestWithParams(ctx, http.MethodGet, c.buildURL("setPower"), params)
	if err != nil {
		return err
	}
	// #nosec G704
	res, err := c.Do(req)
	if err != nil {
		return err
	}
	err = parseHTTPResponse(res, "PowerOn")
	if err != nil {
		return err
	}
	return nil
}

// PowerOff sets the Yamaha device to standby mode.
func (c Client) PowerOff(ctx context.Context) error {
	params := map[string]string{}
	params["power"] = "standby"
	req, err := internal.BuildHttpRequestWithParams(ctx, http.MethodGet, c.buildURL("setPower"), params)
	if err != nil {
		return err
	}
	// #nosec G704
	res, err := c.Do(req)
	if err != nil {
		return err
	}
	err = parseHTTPResponse(res, "PowerOff")
	if err != nil {
		return err
	}
	return nil
}

// SetInput sets the input source on the Yamaha device.
func (c Client) SetInput(ctx context.Context, input string) error {
	params := map[string]string{}
	params["input"] = input
	req, err := internal.BuildHttpRequestWithParams(ctx, http.MethodGet, c.buildURL("setInput"), params)
	if err != nil {
		return err
	}
	// #nosec G704
	res, err := c.Do(req)
	if err != nil {
		return err
	}
	err = parseHTTPResponse(res, "SetInput")
	if err != nil {
		return err
	}
	return nil
}
