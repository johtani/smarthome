package slack

import (
	"os"
	"testing"
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
				AppToken: "xapp-123",
				BotToken: "xoxb-123",
			},
			wantErr: false,
		},
		{
			name: "empty app token",
			config: Config{
				AppToken: "",
				BotToken: "xoxb-123",
			},
			wantErr: true,
		},
		{
			name: "invalid app token prefix",
			config: Config{
				AppToken: "xoxb-123",
				BotToken: "xoxb-123",
			},
			wantErr: true,
		},
		{
			name: "empty bot token",
			config: Config{
				AppToken: "xapp-123",
				BotToken: "",
			},
			wantErr: true,
		},
		{
			name: "invalid bot token prefix",
			config: Config{
				AppToken: "xapp-123",
				BotToken: "xapp-123",
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
	_ = os.Setenv("SMARTHOME_SLACK_APP_TOKEN", "xapp-env-token")
	_ = os.Setenv("SMARTHOME_SLACK_BOT_TOKEN", "xoxb-env-token")
	defer func() {
		_ = os.Unsetenv("SMARTHOME_SLACK_APP_TOKEN")
		_ = os.Unsetenv("SMARTHOME_SLACK_BOT_TOKEN")
	}()

	config := Config{}
	config.overrideWithEnv()

	if config.AppToken != "xapp-env-token" {
		t.Errorf("expected xapp-env-token, got %s", config.AppToken)
	}
	if config.BotToken != "xoxb-env-token" {
		t.Errorf("expected xoxb-env-token, got %s", config.BotToken)
	}
}

func TestConfig_OverrideWithEnv_NoEnv(t *testing.T) {
	_ = os.Unsetenv("SMARTHOME_SLACK_APP_TOKEN")
	_ = os.Unsetenv("SMARTHOME_SLACK_BOT_TOKEN")

	config := Config{AppToken: "xapp-original", BotToken: "xoxb-original"}
	config.overrideWithEnv()

	if config.AppToken != "xapp-original" {
		t.Errorf("expected xapp-original, got %s", config.AppToken)
	}
	if config.BotToken != "xoxb-original" {
		t.Errorf("expected xoxb-original, got %s", config.BotToken)
	}
}
