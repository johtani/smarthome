package switchbot

import (
	"context"
	"reflect"
	"testing"

	"github.com/nasa9084/go-switchbot/v3"
)

func TestNewClient(t *testing.T) {
	config := Config{
		Token:  "test-token",
		Secret: "test-secret",
	}
	client := NewClient(config)

	if client.Client == nil {
		t.Fatal("NewClient() client.Client is nil")
	}

	// check if it is *switchbot.Client
	if reflect.TypeOf(client.Client) != reflect.TypeOf(&switchbot.Client{}) {
		t.Errorf("NewClient() client.Client type is %T, want *switchbot.Client", client.Client)
	}

	// switchbot.Client uses an internal httpClient.
	// We want to check if the transport is otelhttp.Transport.
	// Since it's internal, we might have to use reflection if we really want to check it,
	// but switchbot.Client doesn't export the httpClient.
	// However, we can check the caches are initialized.
	if client.deviceNameCache == nil {
		t.Error("NewClient() deviceNameCache is nil")
	}
	if client.sceneNameCache == nil {
		t.Error("NewClient() sceneNameCache is nil")
	}
}

func TestCachedClient_GetSceneName(t *testing.T) {
	c := NewClient(Config{Token: "token", Secret: "secret"})
	c.sceneNameCache["id1"] = "name1"

	name, err := c.GetSceneName(context.Background(), "id1")
	if err != nil {
		t.Errorf("GetSceneName() error = %v", err)
	}
	if name != "name1" {
		t.Errorf("GetSceneName() = %v, want %v", name, "name1")
	}
}

func TestCachedClient_GetDeviceName(t *testing.T) {
	c := NewClient(Config{Token: "token", Secret: "secret"})
	c.deviceNameCache["id2"] = "name2"

	name, err := c.GetDeviceName(context.Background(), "id2")
	if err != nil {
		t.Errorf("GetDeviceName() error = %v", err)
	}
	if name != "name2" {
		t.Errorf("GetDeviceName() = %v, want %v", name, "name2")
	}
}

func TestIsTargetDevice(t *testing.T) {
	tests := []struct {
		name        string
		deviceTypes []string
		target      string
		want        bool
	}{
		{"match", []string{"Light", "Switch"}, "Light", true},
		{"no match", []string{"Light", "Switch"}, "Air Conditioner", false},
		{"empty list", []string{}, "Light", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsTargetDevice(tt.deviceTypes, tt.target); got != tt.want {
				t.Errorf("IsTargetDevice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	type fields struct {
		token  string
		secret string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{"no args", fields{}, true},
		{"only Token", fields{token: "Token"}, true},
		{"only Secret", fields{secret: "Secret"}, true},
		{"ok Token and Secret", fields{token: "Token", secret: "Secret"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Config{
				Token:  tt.fields.token,
				Secret: tt.fields.secret,
			}
			if err := c.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
