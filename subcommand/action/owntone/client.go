package owntone

import (
	"encoding/json"
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

func (c Client) GetPlaylists() ([]string, error) {
	type Item struct {
		Uri string `json:"uri"`
	}
	type Playlists struct {
		Items []Item `json:"items"`
		Total int    `json:"total"`
	}
	req, err := http.NewRequest(http.MethodGet, c.buildUrl("api/library/playlists"), nil)
	if err != nil {
		return nil, err
	}
	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("something wrong... status code is %d. %v", res.StatusCode, res.Header)
	}
	var p Playlists
	if err := json.NewDecoder(res.Body).Decode(&p); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}
	var lists []string
	for _, item := range p.Items {
		lists = append(lists, item.Uri)
	}
	return lists, nil
}

func (c Client) AddItem2Queue(uri string) error {
	req, err := http.NewRequest(http.MethodPost, c.buildUrl("api/queue/items/add"), nil)
	if err != nil {
		return err
	}
	q := req.URL.Query()
	q.Set("uris", uri)
	req.URL.RawQuery = q.Encode()
	res, err := c.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("something wrong... status code is %d. %v", res.StatusCode, res.Header)
	}
	return nil
}

type PlayerStatus struct {
	State          string `json:"state"`
	Repeat         string `json:"repeat"`
	Consume        bool   `json:"consume"`
	Shuffle        bool   `json:"shuffle"`
	Volume         int    `json:"volume"`
	ItemID         int    `json:"item_id"`
	ItemLengthMS   int    `json:"item_length_ms"`
	ItemProgressMS int    `json:"item_progress_ms"`
}

func (c Client) GetPlayerStatus() (*PlayerStatus, error) {
	req, err := http.NewRequest(http.MethodGet, c.buildUrl("api/player"), nil)
	if err != nil {
		return nil, err
	}
	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("something wrong... status code is %d. %v", res.StatusCode, res.Header)
	}
	var p PlayerStatus
	if err := json.NewDecoder(res.Body).Decode(&p); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}
	return &p, nil
}
