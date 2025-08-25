package owntone

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/johtani/smarthome/subcommand/action/internal"
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
		config: config,
		Client: http.Client{Timeout: 10 * time.Second},
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
	defer func(Body io.ReadCloser) {
		_, _ = io.Copy(io.Discard, Body)
		_ = Body.Close()
	}(res.Body)
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
	defer func(Body io.ReadCloser) {
		_, _ = io.Copy(io.Discard, Body)
		_ = Body.Close()
	}(res.Body)
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
	defer func() {
		_, _ = io.Copy(io.Discard, res.Body)
		_ = res.Body.Close()
	}()
	if res.StatusCode != http.StatusNoContent {
		return fmt.Errorf("something wrong... status code is %d. %v", res.StatusCode, res.Header)
	}
	return nil
}

type Playlist struct {
	Uri       string `json:"uri"`
	Name      string `json:"name"`
	ItemCount int    `json:"item_count"`
	Path      string `json:"path"`
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
	defer func(Body io.ReadCloser) {
		_, _ = io.Copy(io.Discard, Body)
		_ = Body.Close()
	}(res.Body)
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

func (c Client) AddItem2QueueAndPlay(uri string, expression string) error {
	params := map[string]string{"playback": "start"}
	if len(uri) > 0 {
		params["uris"] = uri
	} else if len(expression) > 0 {
		params["expression"] = expression
	}
	req, err := internal.BuildHttpRequestWithParams(http.MethodPost, c.buildUrl("api/queue/items/add"), params)
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
	defer func(Body io.ReadCloser) {
		_, _ = io.Copy(io.Discard, Body)
		_ = Body.Close()
	}(res.Body)
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("something wrong... status code is %d. %v", res.StatusCode, res.Header)
	}
	var p PlayerStatus
	if err := json.NewDecoder(res.Body).Decode(&p); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}
	return &p, nil
}

func (c Client) ClearQueue() error {
	req, err := http.NewRequest(http.MethodPut, c.buildUrl("api/queue/clear"), nil)
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
	if res.StatusCode != http.StatusNoContent {
		return fmt.Errorf("something wrong... status code is %d. %v", res.StatusCode, res.Header)
	}
	return nil
}

type SearchType string

func SearchTypeFromString(s string) (SearchType, error) {

	switch s {
	case "artist":
		return artist, nil
	case "album":
		return album, nil
	case "track":
		return track, nil
	case "genre":
		return genre, nil
	default:
		return artist, fmt.Errorf("not found %v", s)
	}
}

const (
	//playlist SearchType = "playlist"
	artist SearchType = "artist"
	album  SearchType = "album"
	track  SearchType = "track"
	genre  SearchType = "genre" // after https://github.com/owntone/owntone-server/commit/3e7e03b4c18b091b01b66e62467067e7cbf50da4
)

type SearchItem struct {
	Title  string `json:"title"`
	Uri    string `json:"uri"`
	Name   string `json:"name"`
	Artist string `json:"artist"`
}

type Items struct {
	Items  []SearchItem `json:"items"`
	Total  int          `json:"total"`
	Offset int          `json:"offset"`
	Limit  int          `json:"limit"`
}

type SearchResult struct {
	Tracks    Items `json:"tracks"`
	Artists   Items `json:"artists"`
	Albums    Items `json:"albums"`
	Genres    Items `json:"genres"`
	Playlists Items `json:"playlists"`
}

func (c Client) Search(keyword string, resultType []SearchType, limit int) (*SearchResult, error) {
	params := map[string]string{}
	params["query"] = keyword
	l := limit
	if limit <= 0 {
		l = 5
	}
	params["limit"] = strconv.Itoa(l)
	var types []string
	for _, s := range resultType {
		types = append(types, string(s))
	}
	params["type"] = strings.Join(types, ",")
	req, err := internal.BuildHttpRequestWithParams(http.MethodGet, c.buildUrl("api/search"), params)
	if err != nil {
		return nil, err
	}
	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_, _ = io.Copy(io.Discard, Body)
		_ = Body.Close()
	}(res.Body)
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("something wrong... status code is %d. %v", res.StatusCode, res.Header)
	}

	// Decode the JSON response into a Response struct
	var results SearchResult
	if err := json.NewDecoder(res.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}
	return &results, nil
}

type Counts struct {
	Songs   int `json:"songs"`
	Artists int `json:"artists"`
	Albums  int `json:"albums"`
}

func (c Client) Counts() (*Counts, error) {
	req, err := http.NewRequest(http.MethodGet, c.buildUrl("api/library"), nil)
	if err != nil {
		return nil, err
	}
	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_, _ = io.Copy(io.Discard, Body)
		_ = Body.Close()
	}(res.Body)
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("something wrong... status code is %d. %v", res.StatusCode, res.Header)
	}
	var p Counts
	if err := json.NewDecoder(res.Body).Decode(&p); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}
	return &p, nil
}

type Artists struct {
	Items []Artist `json:"items"`
}

type Artist struct {
	Name       string `json:"name"`
	Uri        string `json:"uri"`
	TrackCount int    `json:"track_count"`
}

func (c Client) GetArtist(offset int) (*Artist, error) {

	params := map[string]string{}
	params["offset"] = strconv.Itoa(offset)
	params["limit"] = strconv.Itoa(1)
	req, err := internal.BuildHttpRequestWithParams(http.MethodGet, c.buildUrl("api/library/artists"), params)
	if err != nil {
		return nil, err
	}
	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_, _ = io.Copy(io.Discard, Body)
		_ = Body.Close()
	}(res.Body)
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("something wrong... status code is %d. %v", res.StatusCode, res.Header)
	}

	// Decode the JSON response into a Response struct
	var results Artists
	if err := json.NewDecoder(res.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}
	//
	return &results.Items[0], nil
}

type Genres struct {
	Items []Genre `json:"items"`
}
type Genre struct {
	Name       string `json:"name"`
	TrackCount int    `json:"track_count"`
}

func (c Client) GetGenres() ([]Genre, error) {
	params := map[string]string{}
	req, err := internal.BuildHttpRequestWithParams(http.MethodGet, c.buildUrl("api/library/genres"), params)
	if err != nil {
		return nil, err
	}
	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_, _ = io.Copy(io.Discard, Body)
		_ = Body.Close()
	}(res.Body)
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("something wrong... status code is %d. %v", res.StatusCode, res.Header)
	}

	// Decode the JSON response into a Response struct
	var results Genres
	if err := json.NewDecoder(res.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}
	//
	return results.Items, nil
}

type Output struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Selected bool   `json:"selected"`
	Volume   int    `json:"volume"`
}

type Outputs struct {
	Outputs []Output `json:"outputs"`
}

// GetOutputs fetches the list of audio outputs (speakers) from Owntone.
func (c Client) GetOutputs() ([]Output, error) {
	req, err := http.NewRequest(http.MethodGet, c.buildUrl("api/outputs"), nil)
	if err != nil {
		return nil, err
	}
	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_, _ = io.Copy(io.Discard, Body)
		_ = Body.Close()
	}(res.Body)
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("something wrong... status code is %d. %v", res.StatusCode, res.Header)
	}
	// Decode the JSON response into an Outputs struct
	var results Outputs
	if err := json.NewDecoder(res.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}
	return results.Outputs, nil
}

func (c Client) UpdateLibrary() error {
	req, err := internal.BuildHttpRequestWithParams(http.MethodPut, c.buildUrl("api/update"), nil)
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
	if res.StatusCode != http.StatusNoContent {
		return fmt.Errorf("something wrong... status code is %d. %v", res.StatusCode, res.Header)
	}
	return nil
}
