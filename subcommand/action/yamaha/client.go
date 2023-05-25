package yamaha

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const EnvUrl = "YAMAHA_AMP_URL"
const basePath = "YamahaExtendedControl/v1/main/"

type Client struct {
	config Config
	http.Client
}

type Config struct {
	url string
}

func NewConfig(url string) (Config, error) {
	if len(url) == 0 {
		return Config{}, fmt.Errorf("not found \"YAMAHA_AMP_URL\". Please set YAMAHA_AMP_URL via Environment variable")
	}
	if strings.HasSuffix(url, "/") {
		return Config{
			url,
		}, nil
	}
	return Config{
		url + "/",
	}, nil
}

func (c Client) buildUrl(path string) string {
	return c.config.url + basePath + path
}

func NewClient(config Config) *Client {
	return &Client{
		config,
		http.Client{Timeout: 10 * time.Second},
	}
}

type ResponseCode struct {
	ResponseCode int `json:"response_code"`
}

func buildHttpRequest(method string, url string, params map[string]string) (*http.Request, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	for key, param := range params {
		q.Set(key, param)
	}
	req.URL.RawQuery = q.Encode()
	return req, nil
}

func (c Client) SetScene(scene int) error {
	url := c.buildUrl("recallScene")
	method := http.MethodGet
	params := map[string]string{}
	params["num"] = strconv.Itoa(scene)
	req, err := buildHttpRequest(method, url, params)
	if err != nil {
		return err
	}
	res, err := c.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("something wrong... status code is %d. %v", res.StatusCode, res.Header)
	}
	var rc ResponseCode
	if err := json.NewDecoder(res.Body).Decode(&rc); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}
	if rc.ResponseCode != 0 {
		return fmt.Errorf("something wrong SetScene... response_code is %v", rc.ResponseCode)
	}
	return nil
}

func (c Client) SetVolume(volume int) error {
	url := c.buildUrl("setVolume")
	method := http.MethodGet
	params := map[string]string{}
	params["volume"] = strconv.Itoa(volume)
	req, err := buildHttpRequest(method, url, params)
	if err != nil {
		return err
	}
	res, err := c.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("something wrong... status code is %d. %v", res.StatusCode, res.Header)
	}
	var rc ResponseCode
	if err := json.NewDecoder(res.Body).Decode(&rc); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}
	if rc.ResponseCode != 0 {
		return fmt.Errorf("something wrong SetVolume... response_code is %v", rc.ResponseCode)
	}
	return nil
}
