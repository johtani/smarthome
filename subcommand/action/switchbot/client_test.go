package switchbot

import (
	"testing"
)

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
		{"no args", fields{"", ""}, true},
		{"only Token", fields{"Token", ""}, true},
		{"only Secret", fields{"", "Secret"}, true},
		{"ok Token and Secret", fields{"Token", "Secret"}, false},
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
