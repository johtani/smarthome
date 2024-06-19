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

func TestClient_AddItem2QueueAndPlay(t *testing.T) {
	type fields struct {
		statusCode int
		method     string
		path       string
		uris       string
		expression string
	}
	path := "/api/queue/items/add"
	tests := []struct {
		name           string
		fields         fields
		expectedParams map[string][]string
		wantErr        bool
	}{
		{"uris OK", fields{statusCode: http.StatusOK, method: http.MethodPost, path: path, uris: "playlist", expression: ""}, map[string][]string{"uris": {"playlist"}, "playback": {"start"}}, false},
		{"expression OK", fields{statusCode: http.StatusOK, method: http.MethodPost, path: path, uris: "", expression: "expression"}, map[string][]string{"expression": {"expression"}, "playback": {"start"}}, false},
		{"uris and expressions OK", fields{statusCode: http.StatusOK, method: http.MethodPost, path: path, uris: "playlist", expression: "expression"}, map[string][]string{"uris": {"playlist"}, "playback": {"start"}}, false},
		{"NG", fields{statusCode: http.StatusInternalServerError, method: http.MethodPost, path: path, uris: "playlist", expression: ""}, map[string][]string{"uris": {"playlist"}, "playback": {"start"}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := createMockServer(tt.fields.statusCode, tt.fields.method, tt.fields.path, tt.expectedParams)
			defer server.Close()
			config := Config{Url: server.URL}
			c := NewClient(config)
			if err := c.AddItem2QueueAndPlay(tt.fields.uris, tt.fields.expression); (err != nil) != tt.wantErr {
				t.Errorf("AddItem2QueueAndPlay() error = %v, wantErr %v", err, tt.wantErr)
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
		{"OK", fields{statusCode: http.StatusOK, method: http.MethodGet, path: path, response: playlistsSampleJSONResponse()}, false, []Playlist{{"library:playlist:1", "radio", 491, "/music/srv/radio.m3u"}}},
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
			}
			if !reflect.DeepEqual(status, tt.expected) {
				t.Errorf("GetPlayerStatus() count = %v, want %v", status, tt.expected)
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
  "genres": {
    "items": [
      {
        "name": "Abstract",
        "name_sort": "Abstract",
        "track_count": 3,
        "album_count": 1,
        "artist_count": 1,
        "length_ms": 938426,
        "time_added": "2022-11-17T09:28:08Z",
        "in_progress": false,
        "media_kind": "music",
        "data_kind": "file",
        "year": 0
      },
      {
        "name": "Alternative",
        "name_sort": "Alternative",
        "track_count": 261,
        "album_count": 27,
        "artist_count": 17,
        "length_ms": 61207056,
        "time_played": "2024-02-22T03:13:27Z",
        "time_added": "2022-11-17T09:28:18Z",
        "in_progress": false,
        "media_kind": "music",
        "data_kind": "file",
        "date_released": "2018-01-01",
        "year": 2018
      }
    ],
    "total": 182,
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
				Genres: Items{Items: []SearchItem{
					{
						Title:  "",
						Uri:    "",
						Name:   "Abstract",
						Artist: "",
					},
					{
						Title:  "",
						Uri:    "",
						Name:   "Alternative",
						Artist: "",
					},
				}, Total: 182, Offset: 0, Limit: 2},
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

func countsSampleJSONResponse() string {
	return `{
  "songs": 15942,
  "db_playtime": 3914217,
  "artists": 767,
  "albums": 2256,
  "file_size": 168813164331,
  "started_at": "2024-02-27T14:51:41Z",
  "updated_at": "2024-03-05T08:21:35Z",
  "updating": false,
  "scanners": [
    {
      "name": "files"
    },
    {
      "name": "spotify"
    },
    {
      "name": "rss"
    }
  ]
}
`
}

func TestClient_Counts(t *testing.T) {
	type fields struct {
		statusCode int
		method     string
		path       string
		response   string
	}

	path := "/api/library"
	tests := []struct {
		name     string
		fields   fields
		wantErr  bool
		expected *Counts
	}{
		{"OK", fields{statusCode: http.StatusOK, method: http.MethodGet, path: path, response: countsSampleJSONResponse()}, false, &Counts{Songs: 15942, Artists: 767, Albums: 2256}},
		{"NG", fields{statusCode: http.StatusInternalServerError, method: http.MethodGet, path: path, response: countsSampleJSONResponse()}, true, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := createMockServerWithResponse(tt.fields.statusCode, tt.fields.method, tt.fields.path, nil, tt.fields.response)
			defer server.Close()
			config := Config{Url: server.URL}
			c := NewClient(config)

			count, err := c.Counts()
			if (err != nil) != tt.wantErr {
				t.Errorf("Count() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(count, tt.expected) {
				t.Errorf("Count() count = %v, want %v", count, tt.expected)
			}
		})
	}
}

func getArtstsSampleJSONResponse() string {
	return `{
  "items": [
    {
      "id": "5132191696218976531",
      "name": "Ace Of Base",
      "name_sort": "Ace Of Base",
      "album_count": 4,
      "track_count": 58,
      "length_ms": 12800561,
      "time_played": "2024-02-01T03:09:40Z",
      "time_added": "2023-02-03T16:04:26Z",
      "in_progress": false,
      "media_kind": "music",
      "data_kind": "file",
      "uri": "library:artist:5132191696218976531",
      "artwork_url": ".\/artwork\/group\/708"
    }
  ],
  "total": 767,
  "offset": 2,
  "limit": 1
}
`
}

func TestClient_GetArtist(t *testing.T) {
	type fields struct {
		statusCode int
		method     string
		path       string
		response   string
	}
	type args struct {
		offset int
	}
	path := "/api/library/artists"
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantErr  bool
		expected *Artist
	}{
		{"OK", fields{statusCode: http.StatusOK, method: http.MethodGet, path: path, response: getArtstsSampleJSONResponse()}, args{offset: 1}, false, &Artist{Name: "Ace Of Base", Uri: "library:artist:5132191696218976531", TrackCount: 58}},
		{"NG", fields{statusCode: http.StatusInternalServerError, method: http.MethodGet, path: path, response: getArtstsSampleJSONResponse()}, args{offset: 1}, true, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := createMockServerWithResponse(tt.fields.statusCode, tt.fields.method, tt.fields.path, nil, tt.fields.response)
			defer server.Close()
			config := Config{Url: server.URL}
			c := NewClient(config)

			artist, err := c.GetArtist(tt.args.offset)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetArtist() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(artist, tt.expected) {
				t.Errorf("GetArtist() artist = %v, want %v", artist, tt.expected)
			}
		})
	}
}

func getGenresSampleJSONResponse() string {
	return `{
  "items": [
    {
      "name": "Abstract",
      "name_sort": "Abstract",
      "track_count": 3,
      "album_count": 1,
      "artist_count": 1,
      "length_ms": 938426,
      "time_added": "2022-11-17T09:28:08Z",
      "in_progress": false,
      "media_kind": "music",
      "data_kind": "file",
      "year": 0
    },
    {
      "name": "Alternative",
      "name_sort": "Alternative",
      "track_count": 261,
      "album_count": 27,
      "artist_count": 17,
      "length_ms": 61207056,
      "time_played": "2024-02-22T03:13:27Z",
      "time_added": "2022-11-17T09:28:18Z",
      "in_progress": false,
      "media_kind": "music",
      "data_kind": "file",
      "date_released": "2018-01-01",
      "year": 2018
    }
  ],
  "total": 2,
  "offset": 0,
  "limit": 2
}
`
}

func TestClient_GetGenres(t *testing.T) {
	type fields struct {
		statusCode int
		method     string
		path       string
		response   string
	}
	path := "/api/library/genres"
	tests := []struct {
		name     string
		fields   fields
		wantErr  bool
		expected []Genre
	}{
		{"OK", fields{statusCode: http.StatusOK, method: http.MethodGet, path: path, response: getGenresSampleJSONResponse()}, false, []Genre{{Name: "Abstract", TrackCount: 3}, {Name: "Alternative", TrackCount: 261}}},
		{"NG", fields{statusCode: http.StatusInternalServerError, method: http.MethodGet, path: path, response: getGenresSampleJSONResponse()}, true, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := createMockServerWithResponse(tt.fields.statusCode, tt.fields.method, tt.fields.path, nil, tt.fields.response)
			defer server.Close()
			config := Config{Url: server.URL}
			c := NewClient(config)

			genres, err := c.GetGenres()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetGenres() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(genres, tt.expected) {
				t.Errorf("GetGenres() genres = %v, want %v", genres, tt.expected)
			}
		})
	}
}
