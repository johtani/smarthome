package owntone

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"testing"
	"time"
)

func TestCheckConfig(t *testing.T) {
	tests := []struct {
		name    string
		setenv  func(t *testing.T)
		wantErr bool
	}{
		{"no env", func(t *testing.T) {}, true},
		{"ok", func(t *testing.T) { t.Setenv(EnvUrl, "aaa") }, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setenv(t)
			if err := CheckConfig(); (err != nil) != tt.wantErr {
				t.Errorf("CheckConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_buildUrl(t *testing.T) {
	type fields struct {
		url string
	}
	type args struct {
		path string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{"ok url path", fields{"URL"}, args{"path"}, "URLpath"},
		{"ok only url", fields{"URL"}, args{}, "URL"},
		{"ok only path", fields{}, args{"path"}, "path"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Client{
				url: tt.fields.url,
			}
			if got := c.buildUrl(tt.args.path); got != tt.want {
				t.Errorf("buildUrl() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewOwntoneClient(t *testing.T) {
	tests := []struct {
		name   string
		setenv func(t *testing.T)
		want   Client
	}{
		{"no env", func(t *testing.T) {}, Client{"/", http.Client{Timeout: 10 * time.Second}}},
		{"no slash", func(t *testing.T) { t.Setenv(EnvUrl, "url") }, Client{"url/", http.Client{Timeout: 10 * time.Second}}},
		{"end with slash", func(t *testing.T) { t.Setenv(EnvUrl, "url/") }, Client{"url/", http.Client{Timeout: 10 * time.Second}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setenv(t)
			if got := NewOwntoneClient(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewOwntoneClient() = %v, want %v", got, tt.want)
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
				rw.Write([]byte(response))
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
		{"OK", fields{http.StatusNoContent, http.MethodPut, path}, false},
		{"NG", fields{http.StatusInternalServerError, http.MethodPut, path}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := createMockServer(tt.fields.statusCode, tt.fields.method, tt.fields.path, nil)
			defer server.Close()
			t.Setenv(EnvUrl, server.URL)
			c := NewOwntoneClient()
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
		{"OK", fields{http.StatusNoContent, http.MethodPut, path}, false},
		{"NG", fields{http.StatusInternalServerError, http.MethodPut, path}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := createMockServer(tt.fields.statusCode, tt.fields.method, tt.fields.path, nil)
			defer server.Close()
			t.Setenv(EnvUrl, server.URL)
			c := NewOwntoneClient()
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
		{"OK", fields{http.StatusNoContent, http.MethodPut, path, 33}, false},
		{"NG", fields{http.StatusInternalServerError, http.MethodPut, path, 33}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqParams := map[string][]string{"volume": {strconv.Itoa(tt.fields.volume)}}
			server := createMockServer(tt.fields.statusCode, tt.fields.method, tt.fields.path, reqParams)
			defer server.Close()
			t.Setenv(EnvUrl, server.URL)
			c := NewOwntoneClient()
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
		{"OK", fields{http.StatusOK, http.MethodPost, path, "playlist"}, false},
		{"NG", fields{http.StatusInternalServerError, http.MethodPost, path, "playlist"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqParams := map[string][]string{"uris": {tt.fields.item}}
			server := createMockServer(tt.fields.statusCode, tt.fields.method, tt.fields.path, reqParams)
			defer server.Close()
			t.Setenv(EnvUrl, server.URL)
			c := NewOwntoneClient()
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
      "uri": "library:playlist:1"
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
		expected []string
	}{
		{"OK", fields{http.StatusOK, http.MethodGet, path, playlistsSampleJSONResponse()}, false, []string{"library:playlist:1"}},
		{"NG", fields{http.StatusInternalServerError, http.MethodGet, path, playlistsSampleJSONResponse()}, true, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := createMockServerWithResponse(tt.fields.statusCode, tt.fields.method, tt.fields.path, nil, tt.fields.response)
			defer server.Close()
			t.Setenv(EnvUrl, server.URL)
			c := NewOwntoneClient()

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
		{"OK", fields{http.StatusOK, http.MethodGet, path, playerStatusSampleJSONResponse()}, false, &PlayerStatus{"pause", "off", false, false, 50, 0, 0, 0}},
		{"NG", fields{http.StatusInternalServerError, http.MethodGet, path, playerStatusSampleJSONResponse()}, true, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := createMockServerWithResponse(tt.fields.statusCode, tt.fields.method, tt.fields.path, nil, tt.fields.response)
			defer server.Close()
			t.Setenv(EnvUrl, server.URL)
			c := NewOwntoneClient()

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
