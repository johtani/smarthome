package owntone

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

type Client struct {
	config Config
	http.Client
}

type Config struct {
	URL     string `json:"url"`
	Timeout int    `json:"timeout"`
}

func (c Config) Validate() error {
	if c.URL == "" {
		return fmt.Errorf("owntone.url is required")
	}
	return nil
}

func (c Client) buildURL(path string) string {
	url := c.config.URL
	if !strings.HasSuffix(url, "/") {
		url += "/"
	}
	return url + path
}

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

func (c Client) Pause(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.buildURL("api/player/pause"), nil)
	if err != nil {
		return err
	}
	res, err := c.Do(req)
	if err != nil {
		return err
	}
	return internal.HandleResponse(res, http.StatusNoContent)
}

func (c Client) Play(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.buildURL("api/player/play"), nil)
	if err != nil {
		return err
	}
	res, err := c.Do(req)
	if err != nil {
		return err
	}
	return internal.HandleResponse(res, http.StatusNoContent)
}

func (c Client) SetVolume(ctx context.Context, volume int) error {
	params := map[string]string{}
	params["volume"] = strconv.Itoa(volume)
	req, err := internal.BuildHttpRequestWithParams(ctx, http.MethodPut, c.buildURL("api/player/volume"), params)
	if err != nil {
		return err
	}
	res, err := c.Do(req)
	if err != nil {
		return err
	}
	return internal.HandleResponse(res, http.StatusNoContent)
}

type Playlist struct {
	Uri       string `json:"uri"`
	Name      string `json:"name"`
	ItemCount int    `json:"item_count"`
	Path      string `json:"path"`
}

func (c Client) GetPlaylists(ctx context.Context) ([]Playlist, error) {
	type Playlists struct {
		Items []Playlist `json:"items"`
		Total int        `json:"total"`
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.buildURL("api/library/playlists"), nil)
	if err != nil {
		return nil, err
	}
	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	var p Playlists
	if err := internal.DecodeJSONResponse(res, &p, http.StatusOK); err != nil {
		return nil, err
	}
	var lists []Playlist
	for _, item := range p.Items {
		// 曲がないプレイリストは不要なので返さない
		if item.ItemCount > 0 {
			lists = append(lists, item)
		}
	}
	return lists, nil
}

func (c Client) AddItem2QueueAndPlay(ctx context.Context, uri string, expression string) error {
	params := map[string]string{"playback": "start"}
	if len(uri) > 0 {
		params["uris"] = uri
	} else if len(expression) > 0 {
		params["expression"] = expression
	}
	req, err := internal.BuildHttpRequestWithParams(ctx, http.MethodPost, c.buildURL("api/queue/items/add"), params)
	if err != nil {
		return err
	}
	res, err := c.Do(req)
	if err != nil {
		return err
	}
	return internal.HandleResponse(res, http.StatusOK)
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

func (c Client) GetPlayerStatus(ctx context.Context) (*PlayerStatus, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.buildURL("api/player"), nil)
	if err != nil {
		return nil, err
	}
	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	var p PlayerStatus
	if err := internal.DecodeJSONResponse(res, &p, http.StatusOK); err != nil {
		return nil, err
	}
	return &p, nil
}

func (c Client) ClearQueue(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.buildURL("api/queue/clear"), nil)
	if err != nil {
		return err
	}
	res, err := c.Do(req)
	if err != nil {
		return err
	}
	return internal.HandleResponse(res, http.StatusNoContent)
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
	// playlist SearchType = "playlist"
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

func (c Client) Search(ctx context.Context, keyword string, resultType []SearchType, limit int) (*SearchResult, error) {
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
	req, err := internal.BuildHttpRequestWithParams(ctx, http.MethodGet, c.buildURL("api/search"), params)
	if err != nil {
		return nil, err
	}
	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	// Decode the JSON response into a Response struct
	var results SearchResult
	if err := internal.DecodeJSONResponse(res, &results, http.StatusOK); err != nil {
		return nil, err
	}
	return &results, nil
}

type Counts struct {
	Songs   int `json:"songs"`
	Artists int `json:"artists"`
	Albums  int `json:"albums"`
}

func (c Client) Counts(ctx context.Context) (*Counts, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.buildURL("api/library"), nil)
	if err != nil {
		return nil, err
	}
	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	var p Counts
	if err := internal.DecodeJSONResponse(res, &p, http.StatusOK); err != nil {
		return nil, err
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

func (c Client) GetArtist(ctx context.Context, offset int) (*Artist, error) {

	params := map[string]string{}
	params["offset"] = strconv.Itoa(offset)
	params["limit"] = strconv.Itoa(1)
	req, err := internal.BuildHttpRequestWithParams(ctx, http.MethodGet, c.buildURL("api/library/artists"), params)
	if err != nil {
		return nil, err
	}
	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	// Decode the JSON response into a Response struct
	var results Artists
	if err := internal.DecodeJSONResponse(res, &results, http.StatusOK); err != nil {
		return nil, err
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

func (c Client) GetGenres(ctx context.Context) ([]Genre, error) {
	params := map[string]string{}
	req, err := internal.BuildHttpRequestWithParams(ctx, http.MethodGet, c.buildURL("api/library/genres"), params)
	if err != nil {
		return nil, err
	}
	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	// Decode the JSON response into a Response struct
	var results Genres
	if err := internal.DecodeJSONResponse(res, &results, http.StatusOK); err != nil {
		return nil, err
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
func (c Client) GetOutputs(ctx context.Context) ([]Output, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.buildURL("api/outputs"), nil)
	if err != nil {
		return nil, err
	}
	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	// Decode the JSON response into an Outputs struct
	var results Outputs
	if err := internal.DecodeJSONResponse(res, &results, http.StatusOK); err != nil {
		return nil, err
	}
	return results.Outputs, nil
}

func (c Client) UpdateLibrary(ctx context.Context) error {
	req, err := internal.BuildHttpRequestWithParams(ctx, http.MethodPut, c.buildURL("api/update"), nil)
	if err != nil {
		return err
	}
	res, err := c.Do(req)
	if err != nil {
		return err
	}
	return internal.HandleResponse(res, http.StatusNoContent)
}
