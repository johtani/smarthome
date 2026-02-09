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

const DefaultTimeout = 10 * time.Second
const basePath = "YamahaExtendedControl/v1/main/"

type YamahaAPI interface {
	SetScene(ctx context.Context, scene int) error
	SetVolume(ctx context.Context, volume int) error
	PowerOn(ctx context.Context) error
	PowerOff(ctx context.Context) error
	SetInput(ctx context.Context, input string) error
}

type Client struct {
	config Config
	http.Client
}

type Config struct {
	Url string `json:"url"`
}

func (c Config) Validate() error {
	if c.Url == "" {
		return fmt.Errorf("yamaha.url is required")
	}
	return nil
}

func (c Client) buildUrl(path string) string {
	url := c.config.Url
	if !strings.HasSuffix(url, "/") {
		url = url + "/"
	}
	return url + basePath + path
}

func NewClient(config Config) *Client {
	return &Client{
		config: config,
		Client: http.Client{
			Timeout:   DefaultTimeout,
			Transport: otelhttp.NewTransport(http.DefaultTransport),
		},
	}
}

type ResponseCode struct {
	ResponseCode int `json:"response_code"`
}

func parseHttpResponse(res *http.Response, caller string) error {
	var rc ResponseCode
	if err := internal.DecodeJSONResponse(res, &rc, http.StatusOK); err != nil {
		return err
	}
	if rc.ResponseCode != 0 {
		return fmt.Errorf("something wrong %v... response_code is %v", caller, rc.ResponseCode)
	}
	return nil
}

func (c Client) SetScene(ctx context.Context, scene int) error {
	params := map[string]string{}
	params["num"] = strconv.Itoa(scene)
	req, err := internal.BuildHttpRequestWithParams(ctx, http.MethodGet, c.buildUrl("recallScene"), params)
	if err != nil {
		return err
	}
	res, err := c.Do(req)
	if err != nil {
		return err
	}
	err = parseHttpResponse(res, "SetScene")
	if err != nil {
		return err
	}
	return nil
}

func (c Client) SetVolume(ctx context.Context, volume int) error {
	params := map[string]string{}
	params["volume"] = strconv.Itoa(volume)
	req, err := internal.BuildHttpRequestWithParams(ctx, http.MethodGet, c.buildUrl("setVolume"), params)
	if err != nil {
		return err
	}
	res, err := c.Do(req)
	if err != nil {
		return err
	}
	err = parseHttpResponse(res, "SetVolume")
	if err != nil {
		return err
	}
	return nil
}

func (c Client) PowerOn(ctx context.Context) error {
	params := map[string]string{}
	params["power"] = "on"
	req, err := internal.BuildHttpRequestWithParams(ctx, http.MethodGet, c.buildUrl("setPower"), params)
	if err != nil {
		return err
	}
	res, err := c.Do(req)
	if err != nil {
		return err
	}
	err = parseHttpResponse(res, "PowerOn")
	if err != nil {
		return err
	}
	return nil
}

func (c Client) PowerOff(ctx context.Context) error {
	params := map[string]string{}
	params["power"] = "standby"
	req, err := internal.BuildHttpRequestWithParams(ctx, http.MethodGet, c.buildUrl("setPower"), params)
	if err != nil {
		return err
	}
	res, err := c.Do(req)
	if err != nil {
		return err
	}
	err = parseHttpResponse(res, "PowerOff")
	if err != nil {
		return err
	}
	return nil
}

func (c Client) SetInput(ctx context.Context, input string) error {
	params := map[string]string{}
	params["input"] = input
	req, err := internal.BuildHttpRequestWithParams(ctx, http.MethodGet, c.buildUrl("setInput"), params)
	if err != nil {
		return err
	}
	res, err := c.Do(req)
	if err != nil {
		return err
	}
	err = parseHttpResponse(res, "SetInput")
	if err != nil {
		return err
	}
	return nil
}
