package owntone

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const EnvUrl = "OWNTONE_URL"

type Client struct {
	url string
	http.Client
}

func CheckConfig() error {
	url := os.Getenv(EnvUrl)
	if len(url) == 0 {
		return fmt.Errorf("not found \"OWNTONE_URL\". Please set OWNTONE_URL via Environment variable")
	}
	return nil
}

func (c Client) buildUrl(path string) string {
	return c.url + path
}

func NewOwntoneClient() Client {
	url := os.Getenv(EnvUrl)
	if strings.HasSuffix(url, "/") != true {
		url = url + "/"
	}
	return Client{
		url,
		http.Client{Timeout: 10 * time.Second},
	}
}

func (c Client) Pause() error {
	req, err := http.NewRequest(http.MethodPut, c.buildUrl("api/player/pause"), nil)
	if err != nil {
		return err
	}
	res, err := c.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusNoContent {
		return fmt.Errorf("something wrong... status code is %d. %v", res.StatusCode, res.Header)
	}
	return nil
}

func (c Client) Play() error {
	req, err := http.NewRequest(http.MethodPut, c.buildUrl("api/player/play"), nil)
	if err != nil {
		return err
	}
	res, err := c.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusNoContent {
		return fmt.Errorf("something wrong... status code is %d. %v", res.StatusCode, res.Header)
	}
	return nil
}

func (c Client) SetVolume(volume int) error {
	req, err := http.NewRequest(http.MethodPut, c.buildUrl("api/player/volume"), nil)
	if err != nil {
		return err
	}
	q := req.URL.Query()
	q.Set("volume", strconv.Itoa(volume))
	req.URL.RawQuery = q.Encode()
	res, err := c.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusNoContent {
		return fmt.Errorf("something wrong... status code is %d. %v", res.StatusCode, res.Header)
	}
	return nil
}
