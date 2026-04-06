package subcommand

import (
	"testing"

	"github.com/johtani/smarthome/subcommand/action/owntone"
	"github.com/johtani/smarthome/subcommand/action/switchbot"
	"github.com/johtani/smarthome/subcommand/action/yamaha"
)

func TestValidateMacros(t *testing.T) {
	tests := []struct {
		name    string
		macros  []MacroConfig
		wantErr bool
	}{
		{
			name:    "empty macros",
			macros:  []MacroConfig{},
			wantErr: false,
		},
		{
			name: "valid action types",
			macros: []MacroConfig{
				{
					Name: "test",
					Actions: []ActionSpec{
						{Type: "owntone_pause"},
						{Type: "yamaha_power_on"},
						{Type: "switchbot_send_command"},
						{Type: "wait"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "unknown action type",
			macros: []MacroConfig{
				{
					Name: "bad macro",
					Actions: []ActionSpec{
						{Type: "yamaha_pause"},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "multiple macros with one unknown type",
			macros: []MacroConfig{
				{
					Name:    "good",
					Actions: []ActionSpec{{Type: "owntone_pause"}},
				},
				{
					Name:    "bad",
					Actions: []ActionSpec{{Type: "owntone_stopp"}},
				},
			},
			wantErr: true,
		},
		{
			name: "empty actions",
			macros: []MacroConfig{
				{
					Name:    "empty actions macro",
					Actions: []ActionSpec{},
				},
			},
			wantErr: true,
		},
		{
			name: "nil actions",
			macros: []MacroConfig{
				{
					Name: "nil actions macro",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMacros(tt.macros)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateMacros() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestResolveParam(t *testing.T) {
	config := Config{
		Switchbot: switchbot.Config{
			LightDeviceID:    "device-123",
			LightSceneID:     "scene-456",
			AirConditionerID: "ac-789",
		},
	}

	tests := []struct {
		value    string
		expected string
	}{
		{"$LightDeviceID", "device-123"},
		{"$LightSceneID", "scene-456"},
		{"$AirConditionerID", "ac-789"},
		{"literal-value", "literal-value"},
		{"$UnknownRef", "$UnknownRef"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			got := resolveParam(tt.value, config)
			if got != tt.expected {
				t.Errorf("resolveParam(%q) = %q, want %q", tt.value, got, tt.expected)
			}
		})
	}
}

func TestBuildActionsFromSpecs(t *testing.T) {
	config := Config{
		Owntone:   owntone.Config{URL: "http://localhost:8000"},
		Yamaha:    yamaha.Config{URL: "http://localhost:8080"},
		Switchbot: switchbot.Config{Token: "token", Secret: "secret"},
	}

	t.Run("valid specs", func(t *testing.T) {
		specs := []ActionSpec{
			{Type: "owntone_pause"},
			{Type: "yamaha_power_on"},
			{Type: "wait", Params: map[string]string{"seconds": "1"}},
		}
		actions, err := buildActionsFromSpecs(specs, config)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(actions) != 3 {
			t.Errorf("expected 3 actions, got %d", len(actions))
		}
	})

	t.Run("invalid volume param", func(t *testing.T) {
		specs := []ActionSpec{
			{Type: "yamaha_set_volume", Params: map[string]string{"volume": "not-a-number"}},
		}
		_, err := buildActionsFromSpecs(specs, config)
		if err == nil {
			t.Error("expected error for invalid volume")
		}
	})

	t.Run("invalid scene param", func(t *testing.T) {
		specs := []ActionSpec{
			{Type: "yamaha_set_scene", Params: map[string]string{"scene": "abc"}},
		}
		_, err := buildActionsFromSpecs(specs, config)
		if err == nil {
			t.Error("expected error for invalid scene")
		}
	})

	t.Run("invalid wait seconds param", func(t *testing.T) {
		specs := []ActionSpec{
			{Type: "wait", Params: map[string]string{"seconds": "abc"}},
		}
		_, err := buildActionsFromSpecs(specs, config)
		if err == nil {
			t.Error("expected error for invalid seconds")
		}
	})

	t.Run("unknown switchbot command", func(t *testing.T) {
		specs := []ActionSpec{
			{Type: "switchbot_send_command", Params: map[string]string{"device_id": "dev1", "command": "invalid"}},
		}
		_, err := buildActionsFromSpecs(specs, config)
		if err == nil {
			t.Error("expected error for unknown switchbot command")
		}
	})

	t.Run("nil params defaults to empty map", func(t *testing.T) {
		specs := []ActionSpec{
			{Type: "owntone_display_outputs"},
		}
		actions, err := buildActionsFromSpecs(specs, config)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(actions) != 1 {
			t.Errorf("expected 1 action, got %d", len(actions))
		}
	})
}
