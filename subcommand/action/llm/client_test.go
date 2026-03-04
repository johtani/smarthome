package llm

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_Resolve(t *testing.T) {
	mockResponse := struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}{
		Choices: []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		}{
			{
				Message: struct {
					Content string `json:"content"`
				}{
					Content: `{"command": "light on", "args": "", "thought": "user wants to turn on the light"}`,
				},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST request, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected application/json content-type, got %s", r.Header.Get("Content-Type"))
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	config := Config{
		Endpoint: server.URL,
		Model:    "test-model",
	}
	client := NewClient(config)

	resolved, err := client.Resolve(t.Context(), "電気をつけて", "light on: turn on the light")
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	if resolved.Command != "light on" {
		t.Errorf("expected command 'light on', got '%s'", resolved.Command)
	}
	if resolved.Thought == "" {
		t.Error("expected thought to be non-empty")
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				Endpoint: "http://localhost:8080",
				Model:    "gpt-4o",
			},
			wantErr: false,
		},
		{
			name: "missing endpoint",
			config: Config{
				Model: "gpt-4o",
			},
			wantErr: true,
		},
		{
			name: "missing model",
			config: Config{
				Endpoint: "http://localhost:8080",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.config.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
