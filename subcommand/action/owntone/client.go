/*
Package owntone provides actions and a client for controlling Owntone (formerly forked-daapd) servers.
*/
package owntone

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/hbollon/go-edlib"
	"github.com/johtani/smarthome/subcommand/action/internal"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// DefaultTimeout is the default timeout for Owntone API requests.
const DefaultTimeout = 10 * time.Second

// Client is a client for the Owntone API.
type Client struct {
	config Config
	http.Client
}

// Config is the configuration for the Owntone client.
type Config struct {
	URL     string `json:"url"`
	Timeout int    `json:"timeout"`
}

// Validate validates the Owntone configuration.
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

// NewClient creates a new Owntone client with the given configuration.
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

// Pause pauses playback on the Owntone server.
func (c Client) Pause(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.buildURL("api/player/pause"), nil)
	if err != nil {
		return err
	}
	// #nosec G704
	res, err := c.Do(req)
	if err != nil {
		return err
	}
	return internal.HandleResponse(res, http.StatusNoContent)
}

// Play starts or resumes playback on the Owntone server.
func (c Client) Play(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.buildURL("api/player/play"), nil)
	if err != nil {
		return err
	}
	// #nosec G704
	res, err := c.Do(req)
	if err != nil {
		return err
	}
	return internal.HandleResponse(res, http.StatusNoContent)
}

// SetVolume sets the volume level on the Owntone server.
func (c Client) SetVolume(ctx context.Context, volume int) error {
	params := map[string]string{}
	params["volume"] = strconv.Itoa(volume)
	req, err := internal.BuildHttpRequestWithParams(ctx, http.MethodPut, c.buildURL("api/player/volume"), params)
	if err != nil {
		return err
	}
	// #nosec G704
	res, err := c.Do(req)
	if err != nil {
		return err
	}
	return internal.HandleResponse(res, http.StatusNoContent)
}

// Playlist represents a playlist on the Owntone server.
type Playlist struct {
	URI       string `json:"uri"`
	Name      string `json:"name"`
	ItemCount int    `json:"item_count"`
	Path      string `json:"path"`
}

// GetPlaylists fetches the list of playlists from the Owntone server.
func (c Client) GetPlaylists(ctx context.Context) ([]Playlist, error) {
	type Playlists struct {
		Items []Playlist `json:"items"`
		Total int        `json:"total"`
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.buildURL("api/library/playlists"), nil)
	if err != nil {
		return nil, err
	}
	// #nosec G704
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

// AddItem2QueueAndPlay adds an item (by URI or expression) to the queue and starts playback.
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
	// #nosec G704
	res, err := c.Do(req)
	if err != nil {
		return err
	}
	return internal.HandleResponse(res, http.StatusOK)
}

// PlayerStatus represents the current status of the Owntone player.
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

// GetPlayerStatus fetches the current player status from the Owntone server.
func (c Client) GetPlayerStatus(ctx context.Context) (*PlayerStatus, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.buildURL("api/player"), nil)
	if err != nil {
		return nil, err
	}
	// #nosec G704
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

// ClearQueue clears the current playback queue on the Owntone server.
func (c Client) ClearQueue(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.buildURL("api/queue/clear"), nil)
	if err != nil {
		return err
	}
	// #nosec G704
	res, err := c.Do(req)
	if err != nil {
		return err
	}
	return internal.HandleResponse(res, http.StatusNoContent)
}

// SearchType represents the type of search to perform.
type SearchType string

// SearchTypeFromString converts a string to a SearchType.
func SearchTypeFromString(s string) (SearchType, error) {
	st, err := searchTypeFromString(s)
	if err == nil {
		return st, nil
	}
	// Try fuzzy match
	candidates := []string{string(artist), string(album), string(track), string(genre)}
	res, err := edlib.FuzzySearch(s, candidates, edlib.Levenshtein)
	if err == nil && res != "" {
		distance := edlib.LevenshteinDistance(s, res)
		if distance <= 2 {
			return SearchType(res), nil
		}
	}

	return artist, fmt.Errorf("not found %v", s)
}

func searchTypeFromString(s string) (SearchType, error) {
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

// SearchItem represents an item returned in search results.
type SearchItem struct {
	Title  string `json:"title"`
	URI    string `json:"uri"`
	Name   string `json:"name"`
	Artist string `json:"artist"`
}

// Items represents a collection of search items with pagination info.
type Items struct {
	Items  []SearchItem `json:"items"`
	Total  int          `json:"total"`
	Offset int          `json:"offset"`
	Limit  int          `json:"limit"`
}

// SearchResult represents the results of a search operation.
type SearchResult struct {
	Tracks    Items `json:"tracks"`
	Artists   Items `json:"artists"`
	Albums    Items `json:"albums"`
	Genres    Items `json:"genres"`
	Playlists Items `json:"playlists"`
}

// Search performs a search on the Owntone server.
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
	// #nosec G704
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

// Counts represents library statistics from the Owntone server.
type Counts struct {
	Songs   int `json:"songs"`
	Artists int `json:"artists"`
	Albums  int `json:"albums"`
}

// Counts fetches library statistics from the Owntone server.
func (c Client) Counts(ctx context.Context) (*Counts, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.buildURL("api/library"), nil)
	if err != nil {
		return nil, err
	}
	// #nosec G704
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

// Artists represents a collection of artists from the Owntone server.
type Artists struct {
	Items []Artist `json:"items"`
}

// Artist represents an artist in the Owntone library.
type Artist struct {
	Name       string `json:"name"`
	URI        string `json:"uri"`
	TrackCount int    `json:"track_count"`
}

// GetArtist fetches an artist from the Owntone library by offset.
func (c Client) GetArtist(ctx context.Context, offset int) (*Artist, error) {

	params := map[string]string{}
	params["offset"] = strconv.Itoa(offset)
	params["limit"] = strconv.Itoa(1)
	req, err := internal.BuildHttpRequestWithParams(ctx, http.MethodGet, c.buildURL("api/library/artists"), params)
	if err != nil {
		return nil, err
	}
	// #nosec G704
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

// Genres represents a collection of genres from the Owntone server.
type Genres struct {
	Items []Genre `json:"items"`
}

// Genre represents a genre in the Owntone library.
type Genre struct {
	Name       string `json:"name"`
	TrackCount int    `json:"track_count"`
}

// GetGenres fetches the list of genres from the Owntone server.
func (c Client) GetGenres(ctx context.Context) ([]Genre, error) {
	params := map[string]string{}
	req, err := internal.BuildHttpRequestWithParams(ctx, http.MethodGet, c.buildURL("api/library/genres"), params)
	if err != nil {
		return nil, err
	}
	// #nosec G704
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

// Output represents an audio output on the Owntone server.
type Output struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Selected bool   `json:"selected"`
	Volume   int    `json:"volume"`
}

// Outputs represents a collection of audio outputs.
type Outputs struct {
	Outputs []Output `json:"outputs"`
}

// GetOutputs fetches the list of audio outputs (speakers) from Owntone.
// GetOutputs fetches the list of audio outputs from the Owntone server.
func (c Client) GetOutputs(ctx context.Context) ([]Output, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.buildURL("api/outputs"), nil)
	if err != nil {
		return nil, err
	}
	// #nosec G704
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

// UpdateLibrary triggers a library update on the Owntone server.
func (c Client) UpdateLibrary(ctx context.Context) error {
	req, err := internal.BuildHttpRequestWithParams(ctx, http.MethodPut, c.buildURL("api/update"), nil)
	if err != nil {
		return err
	}
	// #nosec G704
	res, err := c.Do(req)
	if err != nil {
		return err
	}
	return internal.HandleResponse(res, http.StatusNoContent)
}
