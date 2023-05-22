package yamaha

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"testing"
)

func TestNewConfig(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name    string
		args    args
		want    Config
		wantErr bool
	}{
		{"OK:end with slash", args{"hogehoge/"}, Config{"hogehoge/"}, false},
		{"OK:end without slash", args{"hogehoge"}, Config{"hogehoge/"}, false},
		{"NG:end without slash", args{""}, Config{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewConfig(tt.args.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewConfig() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_buildUrl(t *testing.T) {
	type fields struct {
		config Config
		Client http.Client
	}
	config, _ := NewConfig("URL")
	type args struct {
		path string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{"ok url path", fields{config, http.Client{}}, args{"path"}, "URL/YamahaExtendedControl/v1/main/path"},
		{"ok only url", fields{config, http.Client{}}, args{}, "URL/YamahaExtendedControl/v1/main/"},
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
		{"OK", fields{http.StatusOK, http.MethodGet, path, 2, responseCodeOK()}, false},
		{"OK_body_NG", fields{http.StatusOK, http.MethodGet, path, 2, responseCodeNG()}, true},
		{"NG", fields{http.StatusInternalServerError, http.MethodGet, path, 2, responseCodeNG()}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqParams := map[string][]string{"num": {strconv.Itoa(tt.fields.scene)}}
			server := createMockServerWithResponse(tt.fields.statusCode, tt.fields.method, tt.fields.path, reqParams, tt.fields.response)
			defer server.Close()
			config, _ := NewConfig(server.URL)
			c := NewClient(config)
			if err := c.SetScene(tt.fields.scene); (err != nil) != tt.wantErr {
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
		{"OK", fields{http.StatusOK, http.MethodGet, path, 2, responseCodeOK()}, false},
		{"OK_body_NG", fields{http.StatusOK, http.MethodGet, path, 2, responseCodeNG()}, true},
		{"NG", fields{http.StatusInternalServerError, http.MethodGet, path, 2, responseCodeNG()}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqParams := map[string][]string{"volume": {strconv.Itoa(tt.fields.volume)}}
			server := createMockServerWithResponse(tt.fields.statusCode, tt.fields.method, tt.fields.path, reqParams, tt.fields.response)
			defer server.Close()
			config, _ := NewConfig(server.URL)
			c := NewClient(config)
			if err := c.SetVolume(tt.fields.volume); (err != nil) != tt.wantErr {
				t.Errorf("SetVolume() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
