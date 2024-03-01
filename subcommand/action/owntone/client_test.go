package owntone

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"testing"
)

func TestClient_buildUrl(t *testing.T) {
	type fields struct {
		config Config
		Client http.Client
	}
	config := Config{Url: "URL"}
	type args struct {
		path string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{"ok Url path", fields{config: config}, args{path: "path"}, "URL/path"},
		{"ok only Url", fields{config: config}, args{}, "URL/"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Client{
				config: tt.fields.config,
				Client: tt.fields.Client,
			}
			if got := c.buildUrl(tt.args.path); got != tt.want {
				t.Errorf("buildUrl() = %v, want %v", got, tt.want)
			}
		})
	}
}

func createMockServer(code int, method string, path string, requestParam map[string][]string) *httptest.Server {
	return createMockServerWithResponse(code, method, path, requestParam, "")
}

func createMockServerWithResponse(code int, method string, path string, requestParam map[string][]string, response string) *httptest.Server {
	return httptest.NewServer(
		http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if req.Method != method {
				rw.WriteHeader(http.StatusBadRequest)
				return
			}
			if req.URL.Path != path {
				rw.WriteHeader(http.StatusBadRequest)
				return
			}
			if requestParam != nil {
				if areMapsEqual(req.URL.Query(), requestParam) != true {
					rw.WriteHeader(http.StatusBadRequest)
					return
				}
			}
			rw.WriteHeader(code)
			if response != "" {
				_, _ = rw.Write([]byte(response))
			}
			return
		}))
}

func areMapsEqual(m1, m2 map[string][]string) bool {
	m1JSON, _ := json.Marshal(m1)
	m2JSON, _ := json.Marshal(m2)
	return string(m1JSON) == string(m2JSON)
}

func TestClient_Pause(t *testing.T) {
	type fields struct {
		statusCode int
		method     string
		path       string
	}
	path := "/api/player/pause"
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{"OK", fields{statusCode: http.StatusNoContent, method: http.MethodPut, path: path}, false},
		{"NG", fields{statusCode: http.StatusInternalServerError, method: http.MethodPut, path: path}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := createMockServer(tt.fields.statusCode, tt.fields.method, tt.fields.path, nil)
			defer server.Close()
			config := Config{Url: server.URL}
			c := NewClient(config)
			if err := c.Pause(); (err != nil) != tt.wantErr {
				t.Errorf("Pause() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_Play(t *testing.T) {
	type fields struct {
		statusCode int
		method     string
		path       string
	}
	path := "/api/player/play"
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{"OK", fields{statusCode: http.StatusNoContent, method: http.MethodPut, path: path}, false},
		{"NG", fields{statusCode: http.StatusInternalServerError, method: http.MethodPut, path: path}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := createMockServer(tt.fields.statusCode, tt.fields.method, tt.fields.path, nil)
			defer server.Close()
			config := Config{Url: server.URL}
			c := NewClient(config)
			if err := c.Play(); (err != nil) != tt.wantErr {
				t.Errorf("Play() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_SetVolume(t *testing.T) {
	type fields struct {
		statusCode int
		method     string
		path       string
		volume     int
	}
	path := "/api/player/volume"
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{"OK", fields{statusCode: http.StatusNoContent, method: http.MethodPut, path: path, volume: 33}, false},
		{"NG", fields{statusCode: http.StatusInternalServerError, method: http.MethodPut, path: path, volume: 33}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqParams := map[string][]string{"volume": {strconv.Itoa(tt.fields.volume)}}
			server := createMockServer(tt.fields.statusCode, tt.fields.method, tt.fields.path, reqParams)
			defer server.Close()
			config := Config{Url: server.URL}
			c := NewClient(config)
			if err := c.SetVolume(tt.fields.volume); (err != nil) != tt.wantErr {
				t.Errorf("SetVolume() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_AddItem2Queue(t *testing.T) {
	type fields struct {
		statusCode int
		method     string
		path       string
		item       string
	}
	path := "/api/queue/items/add"
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{"OK", fields{statusCode: http.StatusOK, method: http.MethodPost, path: path, item: "playlist"}, false},
		{"NG", fields{statusCode: http.StatusInternalServerError, method: http.MethodPost, path: path, item: "playlist"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqParams := map[string][]string{"uris": {tt.fields.item}}
			server := createMockServer(tt.fields.statusCode, tt.fields.method, tt.fields.path, reqParams)
			defer server.Close()
			config := Config{Url: server.URL}
			c := NewClient(config)
			if err := c.AddItem2Queue(tt.fields.item); (err != nil) != tt.wantErr {
				t.Errorf("AddItem2Queue() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func playlistsSampleJSONResponse() string {
	return `{
  "items": [
    {
      "id": 1,
      "name": "radio",
      "path": "/music/srv/radio.m3u",
      "smart_playlist": false,
      "uri": "library:playlist:1",
      "item_count": 491
    },
    {
      "id": 2,
      "name": "stereo",
      "path": "/music/srv/stereo.m3u",
      "smart_playlist": false,
      "uri": "library:playlist:2",
      "item_count": 0
    }
  ],
  "total": 1,
  "offset": 0,
  "limit": -1
}
`
}

func TestClient_GetPlaylists(t *testing.T) {
	type fields struct {
		statusCode int
		method     string
		path       string
		response   string
	}

	path := "/api/library/playlists"
	tests := []struct {
		name     string
		fields   fields
		wantErr  bool
		expected []Playlist
	}{
		{"OK", fields{statusCode: http.StatusOK, method: http.MethodGet, path: path, response: playlistsSampleJSONResponse()}, false, []Playlist{{"library:playlist:1", "radio", 491}}},
		{"NG", fields{statusCode: http.StatusInternalServerError, method: http.MethodGet, path: path, response: playlistsSampleJSONResponse()}, true, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := createMockServerWithResponse(tt.fields.statusCode, tt.fields.method, tt.fields.path, nil, tt.fields.response)
			defer server.Close()
			config := Config{Url: server.URL}
			c := NewClient(config)

			playlists, err := c.GetPlaylists()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPlaylists() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(playlists, tt.expected) {
				t.Errorf("GetPlaylists() playlists = %v, expected %v", playlists, tt.expected)
			}

		})
	}
}

func playerStatusSampleJSONResponse() string {
	return `{
  "state": "pause",
  "repeat": "off",
  "consume": false,
  "shuffle": false,
  "volume": 50,
  "item_id": 0,
  "item_length_ms": 0,
  "item_progress_ms": 0
}
`
}

func TestClient_GetGetPlayerStatus(t *testing.T) {
	type fields struct {
		statusCode int
		method     string
		path       string
		response   string
	}

	path := "/api/player"
	tests := []struct {
		name     string
		fields   fields
		wantErr  bool
		expected *PlayerStatus
	}{
		{"OK", fields{statusCode: http.StatusOK, method: http.MethodGet, path: path, response: playerStatusSampleJSONResponse()}, false, &PlayerStatus{State: "pause", Repeat: "off", Volume: 50}},
		{"NG", fields{statusCode: http.StatusInternalServerError, method: http.MethodGet, path: path, response: playerStatusSampleJSONResponse()}, true, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := createMockServerWithResponse(tt.fields.statusCode, tt.fields.method, tt.fields.path, nil, tt.fields.response)
			defer server.Close()
			config := Config{Url: server.URL}
			c := NewClient(config)

			status, err := c.GetPlayerStatus()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPlayerStatus() error = %v, wantErr %v", err, tt.wantErr)
			} else if err == nil {
				if status == tt.expected {
					t.Errorf("GetPlayerStatus() status = %v, expected %v", status, tt.expected)
				}
			}
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	type fields struct {
		url string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{"OK", fields{url: "Url"}, false},
		{"NG", fields{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Config{
				Url: tt.fields.url,
			}
			if err := c.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_ClearQueue(t *testing.T) {
	type fields struct {
		statusCode int
		method     string
		path       string
	}
	path := "/api/queue/clear"
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{"OK", fields{statusCode: http.StatusNoContent, method: http.MethodPut, path: path}, false},
		{"NG", fields{statusCode: http.StatusInternalServerError, method: http.MethodPut, path: path}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := createMockServer(tt.fields.statusCode, tt.fields.method, tt.fields.path, nil)
			defer server.Close()
			config := Config{Url: server.URL}
			c := NewClient(config)
			if err := c.ClearQueue(); (err != nil) != tt.wantErr {
				t.Errorf("ClearQueue() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func searchSampleJSONResponse() string {
	return `{
  "tracks": {
    "items": [
      {
        "id": 35,
        "title": "Another Love",
        "artist": "Tom Odell",
        "artist_sort": "Tom Odell",
        "album": "Es is was es is",
        "album_sort": "Es is was es is",
        "album_id": "6494853621007413058",
        "album_artist": "Various artists",
        "album_artist_sort": "Various artists",
        "album_artist_id": "8395563705718003786",
        "genre": "Singer/Songwriter",
        "year": 2013,
        "track_number": 7,
        "disc_number": 1,
        "length_ms": 251030,
        "play_count": 0,
        "media_kind": "music",
        "data_kind": "file",
        "path": "/music/srv/Compilations/Es is was es is/07 Another Love.m4a",
        "uri": "library:track:35"
      },
      {
        "id": 215,
        "title": "Away From the Sun",
        "artist": "3 Doors Down",
        "artist_sort": "3 Doors Down",
        "album": "Away From the Sun",
        "album_sort": "Away From the Sun",
        "album_id": "8264078270267374619",
        "album_artist": "3 Doors Down",
        "album_artist_sort": "3 Doors Down",
        "album_artist_id": "5030128490104968038",
        "genre": "Rock",
        "year": 2002,
        "track_number": 2,
        "disc_number": 1,
        "length_ms": 233278,
        "play_count": 0,
        "media_kind": "music",
        "data_kind": "file",
        "path": "/music/srv/Away From the Sun/02 Away From the Sun.mp3",
        "uri": "library:track:215"
      }
    ],
    "total": 14,
    "offset": 0,
    "limit": 2
  },
  "artists": {
    "items": [
      {
        "id": "8737690491750445895",
        "name": "The xx",
        "name_sort": "xx, The",
        "album_count": 2,
        "track_count": 25,
        "length_ms": 5229196,
        "uri": "library:artist:8737690491750445895"
      }
    ],
    "total": 1,
    "offset": 0,
    "limit": 2
  },
  "albums": {
    "items": [
      {
        "id": "8264078270267374619",
        "name": "Away From the Sun",
        "name_sort": "Away From the Sun",
        "artist": "3 Doors Down",
        "artist_id": "5030128490104968038",
        "track_count": 12,
        "length_ms": 2818174,
        "uri": "library:album:8264078270267374619"
      },
      {
        "id": "6835720495312674468",
        "name": "The Better Life",
        "name_sort": "Better Life",
        "artist": "3 Doors Down",
        "artist_id": "5030128490104968038",
        "track_count": 11,
        "length_ms": 2393332,
        "uri": "library:album:6835720495312674468"
      }
    ],
    "total": 3,
    "offset": 0,
    "limit": 2
  },
  "playlists": {
    "items": [],
    "total": 0,
    "offset": 0,
    "limit": 2
  }
}
`
}

func TestClient_Search(t *testing.T) {
	type fields struct {
		statusCode int
		method     string
		path       string
		response   string
	}
	type args struct {
		keyword    string
		resultType []SearchType
		limit      int
	}
	path := "/api/search"
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *SearchResult
		wantErr bool
	}{
		{
			name:   "OK",
			fields: fields{statusCode: http.StatusOK, method: http.MethodGet, path: path, response: searchSampleJSONResponse()},
			args:   args{keyword: "keyword", resultType: []SearchType{track}},
			want: &SearchResult{
				Tracks: Items{Items: []SearchItem{
					{
						Title:  "Another Love",
						Uri:    "library:track:35",
						Name:   "",
						Artist: "Tom Odell",
					},
					{
						Title:  "Away From the Sun",
						Uri:    "library:track:215",
						Name:   "",
						Artist: "3 Doors Down",
					}}, Total: 14, Offset: 0, Limit: 2},
				Artists: Items{Items: []SearchItem{{
					Title:  "",
					Uri:    "library:artist:8737690491750445895",
					Name:   "The xx",
					Artist: "",
				}}, Total: 1, Offset: 0, Limit: 2},
				Albums: Items{Items: []SearchItem{
					{
						Title:  "",
						Uri:    "library:album:8264078270267374619",
						Name:   "Away From the Sun",
						Artist: "3 Doors Down",
					},
					{
						Title:  "",
						Uri:    "library:album:6835720495312674468",
						Name:   "The Better Life",
						Artist: "3 Doors Down",
					},
				}, Total: 3, Offset: 0, Limit: 2},
				Playlists: Items{Items: []SearchItem{}, Total: 0, Offset: 0, Limit: 2},
			},
			wantErr: false,
		},
		{"NG", fields{statusCode: http.StatusInternalServerError, method: http.MethodGet, path: path, response: searchSampleJSONResponse()}, args{keyword: "keyword", resultType: []SearchType{track}}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := createMockServerWithResponse(tt.fields.statusCode, tt.fields.method, tt.fields.path, nil, tt.fields.response)
			defer server.Close()
			config := Config{Url: server.URL}
			c := NewClient(config)

			got, err := c.Search(tt.args.keyword, tt.args.resultType, tt.args.limit)
			if (err != nil) != tt.wantErr {
				t.Errorf("Search() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Search() got = %v, want %v", got, tt.want)
			}
		})
	}
}
