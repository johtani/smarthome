package owntone

import (
	"reflect"
	"testing"
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
		{"no env", func(t *testing.T) {}, Client{"/"}},
		{"no slash", func(t *testing.T) { t.Setenv(EnvUrl, "url") }, Client{"url/"}},
		{"end with slash", func(t *testing.T) { t.Setenv(EnvUrl, "url/") }, Client{"url/"}},
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
