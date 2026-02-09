package subcommand

import (
	"os"
	"testing"

	"github.com/johtani/smarthome/subcommand/action/owntone"
	"github.com/johtani/smarthome/subcommand/action/switchbot"
	"github.com/johtani/smarthome/subcommand/action/yamaha"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				Owntone:   owntone.Config{Url: "http://localhost:8000"},
				Switchbot: switchbot.Config{Token: "token", Secret: "secret"},
				Yamaha:    yamaha.Config{Url: "http://localhost:8080"},
			},
			wantErr: false,
		},
		{
			name: "invalid owntone config",
			config: Config{
				Owntone:   owntone.Config{Url: ""},
				Switchbot: switchbot.Config{Token: "token", Secret: "secret"},
				Yamaha:    yamaha.Config{Url: "http://localhost:8080"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.config.validate(); (err != nil) != tt.wantErr {
				t.Errorf("Config.validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_OverrideWithEnv(t *testing.T) {
	config := Config{}
	os.Setenv("SMARTHOME_OWNTONE_URL", "http://env-owntone")
	os.Setenv("SMARTHOME_SWITCHBOT_TOKEN", "env-token")
	os.Setenv("SMARTHOME_SWITCHBOT_SECRET", "env-secret")
	os.Setenv("SMARTHOME_YAMAHA_URL", "http://env-yamaha")
	os.Setenv("SMARTHOME_INFLUXDB_TOKEN", "env-influx-token")
	os.Setenv("SMARTHOME_INFLUXDB_URL", "http://env-influx-url")
	os.Setenv("SMARTHOME_INFLUXDB_BUCKET", "env-bucket")
	defer func() {
		os.Unsetenv("SMARTHOME_OWNTONE_URL")
		os.Unsetenv("SMARTHOME_SWITCHBOT_TOKEN")
		os.Unsetenv("SMARTHOME_SWITCHBOT_SECRET")
		os.Unsetenv("SMARTHOME_YAMAHA_URL")
		os.Unsetenv("SMARTHOME_INFLUXDB_TOKEN")
		os.Unsetenv("SMARTHOME_INFLUXDB_URL")
		os.Unsetenv("SMARTHOME_INFLUXDB_BUCKET")
	}()

	config.overrideWithEnv()

	if config.Owntone.Url != "http://env-owntone" {
		t.Errorf("expected http://env-owntone, got %s", config.Owntone.Url)
	}
	if config.Switchbot.Token != "env-token" {
		t.Errorf("expected env-token, got %s", config.Switchbot.Token)
	}
	if config.Switchbot.Secret != "env-secret" {
		t.Errorf("expected env-secret, got %s", config.Switchbot.Secret)
	}
	if config.Yamaha.Url != "http://env-yamaha" {
		t.Errorf("expected http://env-yamaha, got %s", config.Yamaha.Url)
	}
	if config.Influxdb.Token != "env-influx-token" {
		t.Errorf("expected env-influx-token, got %s", config.Influxdb.Token)
	}
	if config.Influxdb.Url != "http://env-influx-url" {
		t.Errorf("expected http://env-influx-url, got %s", config.Influxdb.Url)
	}
	if config.Influxdb.Bucket != "env-bucket" {
		t.Errorf("expected env-bucket, got %s", config.Influxdb.Bucket)
	}
}

func TestLoadConfigWithPath(t *testing.T) {
	// 実際の設定ファイルを使うのは避けたいので、一時的なファイルを作成する
	content := `{
		"Owntone": {"url": "http://localhost:8000"},
		"Switchbot": {"token": "token", "secret": "secret"},
		"Yamaha": {"url": "http://localhost:8080"},
		"Influxdb": {"url": "http://localhost:8086", "token": "token", "bucket": "bucket", "org": "org"}
	}`
	tmpfile, err := os.CreateTemp("", "config_test.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	config, err := LoadConfigWithPath(tmpfile.Name())
	if err != nil {
		t.Fatalf("LoadConfigWithPath failed: %v", err)
	}

	if config.Owntone.Url != "http://localhost:8000" {
		t.Errorf("expected http://localhost:8000, got %s", config.Owntone.Url)
	}
	if len(config.Commands.Definitions) == 0 {
		t.Error("commands should be initialized")
	}
}
