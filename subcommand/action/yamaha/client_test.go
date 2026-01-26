package yamaha

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
		{"ok Url path", fields{config: config}, args{path: "path"}, "URL/YamahaExtendedControl/v1/main/path"},
		{"ok only Url", fields{config: config}, args{}, "URL/YamahaExtendedControl/v1/main/"},
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

func responseCodeOK() string {
	return `{
  "response_code": 0
}
`
}

func responseCodeNG() string {
	return `{
  "response_code": 5
}
`
}

func TestClient_SetScene(t *testing.T) {
	type fields struct {
		statusCode int
		method     string
		path       string
		scene      int
		response   string
	}
	path := "/YamahaExtendedControl/v1/main/recallScene"
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{"OK", fields{statusCode: http.StatusOK, method: http.MethodGet, path: path, scene: 2, response: responseCodeOK()}, false},
		{"OK_body_NG", fields{statusCode: http.StatusOK, method: http.MethodGet, path: path, scene: 2, response: responseCodeNG()}, true},
		{"NG", fields{statusCode: http.StatusInternalServerError, method: http.MethodGet, path: path, scene: 2, response: responseCodeNG()}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqParams := map[string][]string{"num": {strconv.Itoa(tt.fields.scene)}}
			server := createMockServerWithResponse(tt.fields.statusCode, tt.fields.method, tt.fields.path, reqParams, tt.fields.response)
			defer server.Close()
			config := Config{Url: server.URL}
			c := NewClient(config)
			if err := c.SetScene(context.Background(), tt.fields.scene); (err != nil) != tt.wantErr {
				t.Errorf("SetScene() error = %v, wantErr %v", err, tt.wantErr)
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
		response   string
	}
	path := "/YamahaExtendedControl/v1/main/setVolume"
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{"OK", fields{statusCode: http.StatusOK, method: http.MethodGet, path: path, volume: 2, response: responseCodeOK()}, false},
		{"OK_body_NG", fields{statusCode: http.StatusOK, method: http.MethodGet, path: path, volume: 2, response: responseCodeNG()}, true},
		{"NG", fields{statusCode: http.StatusInternalServerError, method: http.MethodGet, path: path, volume: 2, response: responseCodeNG()}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqParams := map[string][]string{"volume": {strconv.Itoa(tt.fields.volume)}}
			server := createMockServerWithResponse(tt.fields.statusCode, tt.fields.method, tt.fields.path, reqParams, tt.fields.response)
			defer server.Close()
			config := Config{Url: server.URL}
			c := NewClient(config)
			if err := c.SetVolume(context.Background(), tt.fields.volume); (err != nil) != tt.wantErr {
				t.Errorf("SetVolume() error = %v, wantErr %v", err, tt.wantErr)
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

func TestClient_PowerOff(t *testing.T) {
	type fields struct {
		statusCode int
		method     string
		path       string
		response   string
	}
	path := "/YamahaExtendedControl/v1/main/setPower"
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{"OK", fields{statusCode: http.StatusOK, method: http.MethodGet, path: path, response: responseCodeOK()}, false},
		{"OK_body_NG", fields{statusCode: http.StatusOK, method: http.MethodGet, path: path, response: responseCodeNG()}, true},
		{"NG", fields{statusCode: http.StatusInternalServerError, method: http.MethodGet, path: path, response: responseCodeNG()}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqParams := map[string][]string{"power": {"standby"}}
			server := createMockServerWithResponse(tt.fields.statusCode, tt.fields.method, tt.fields.path, reqParams, tt.fields.response)
			defer server.Close()
			config := Config{Url: server.URL}
			c := NewClient(config)
			if err := c.PowerOff(context.Background()); (err != nil) != tt.wantErr {
				t.Errorf("PowerOff() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
