package yamaha

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/johtani/smarthome/subcommand/action/internal"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const basePath = "YamahaExtendedControl/v1/main/"

type Client struct {
	config Config
	http.Client
}

type Config struct {
	Url string `json:"url"`
}

func (c Config) Validate() error {
	if c.Url == "" {
		return fmt.Errorf("yamaha Url is null")
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
			Timeout:   10 * time.Second,
			Transport: otelhttp.NewTransport(http.DefaultTransport),
		},
	}
}

type ResponseCode struct {
	ResponseCode int `json:"response_code"`
}

func parseHttpResponse(res *http.Response, caller string) error {
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("something wrong... status code is %d. %v", res.StatusCode, res.Header)
	}
	var rc ResponseCode
	if err := json.NewDecoder(res.Body).Decode(&rc); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}
	if rc.ResponseCode != 0 {
		return fmt.Errorf("something wrong %v... response_code is %v", caller, rc.ResponseCode)
	}
	return nil
}

func (c Client) SetScene(scene int) error {
	params := map[string]string{}
	params["num"] = strconv.Itoa(scene)
	req, err := internal.BuildHttpRequestWithParams(context.Background(), http.MethodGet, c.buildUrl("recallScene"), params)
	if err != nil {
		return err
	}
	res, err := c.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		_, _ = io.Copy(io.Discard, Body)
		_ = Body.Close()
	}(res.Body)
	err = parseHttpResponse(res, "SetScene")
	if err != nil {
		return err
	}
	return nil
}

func (c Client) SetVolume(volume int) error {
	params := map[string]string{}
	params["volume"] = strconv.Itoa(volume)
	req, err := internal.BuildHttpRequestWithParams(context.Background(), http.MethodGet, c.buildUrl("setVolume"), params)
	if err != nil {
		return err
	}
	res, err := c.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		_, _ = io.Copy(io.Discard, Body)
		_ = Body.Close()
	}(res.Body)
	err = parseHttpResponse(res, "SetVolume")
	if err != nil {
		return err
	}
	return nil
}

func (c Client) PowerOff() error {
	params := map[string]string{}
	params["power"] = "standby"
	req, err := internal.BuildHttpRequestWithParams(context.Background(), http.MethodGet, c.buildUrl("setPower"), params)
	if err != nil {
		return err
	}
	res, err := c.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		_, _ = io.Copy(io.Discard, Body)
		_ = Body.Close()
	}(res.Body)
	err = parseHttpResponse(res, "PowerOff")
	if err != nil {
		return err
	}
	return nil
}
