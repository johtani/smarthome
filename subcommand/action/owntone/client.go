package owntone

import (
	"encoding/json"
	"fmt"
	"github.com/johtani/smarthome/subcommand/action/internal"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	config Config
	http.Client
}

type Config struct {
	Url string `json:"url"`
}

func (c Config) Validate() error {
	if c.Url == "" {
		return fmt.Errorf("owntone Url is null")
	}
	return nil
}

func (c Client) buildUrl(path string) string {
	url := c.config.Url
	if !strings.HasSuffix(url, "/") {
		url = url + "/"
	}
	return url + path
}

func NewClient(config Config) *Client {
	return &Client{
		config,
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
	params := map[string]string{}
	params["volume"] = strconv.Itoa(volume)
	req, err := internal.BuildHttpRequestWithParams(http.MethodPut, c.buildUrl("api/player/volume"), params)
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

type Playlist struct {
	Uri       string `json:"uri"`
	Name      string `json:"name"`
	ItemCount int    `json:"item_count"`
}

func (c Client) GetPlaylists() ([]Playlist, error) {
	type Playlists struct {
		Items []Playlist `json:"items"`
		Total int        `json:"total"`
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
	var lists []Playlist
	for _, item := range p.Items {
		//曲がないプレイリストは不要なので返さない
		if item.ItemCount > 0 {
			lists = append(lists, item)
		}
	}
	return lists, nil
}

func (c Client) AddItem2Queue(uri string) error {
	params := map[string]string{}
	params["uris"] = uri
	req, err := internal.BuildHttpRequestWithParams(http.MethodPost, c.buildUrl("api/queue/items/add"), params)
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
