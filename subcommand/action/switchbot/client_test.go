package switchbot

import (
	"reflect"
	"testing"
)

func TestNewConfig(t *testing.T) {
	type args struct {
		token  string
		secret string
	}
	tests := []struct {
		name    string
		args    args
		want    Config
		wantErr bool
	}{
		{"no args", args{"", ""}, Config{}, true},
		{"only token", args{"token", ""}, Config{}, true},
		{"only secret", args{"", "secret"}, Config{}, true},
		{"ok token and secret", args{"token", "secret"}, Config{"token", "secret"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewConfig(tt.args.token, tt.args.secret)
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
