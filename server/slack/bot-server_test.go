package slack

import (
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
