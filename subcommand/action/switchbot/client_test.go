package switchbot

import "testing"

func TestCheckConfig(t *testing.T) {
	tests := []struct {
		name    string
		setenv  func(t *testing.T)
		wantErr bool
	}{
		{"no env", func(t *testing.T) {}, true},
		{"only token", func(t *testing.T) { t.Setenv(EnvToken, "token") }, true},
		{"only secret", func(t *testing.T) { t.Setenv(EnvSecret, "secret") }, true},
		{"ok token and secret", func(t *testing.T) { t.Setenv(EnvSecret, "secret"); t.Setenv(EnvToken, "token") }, false},
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
