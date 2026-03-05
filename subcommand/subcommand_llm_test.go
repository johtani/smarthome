package subcommand

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/johtani/smarthome/subcommand/action/llm"
)

func TestCommands_Find_LLM(t *testing.T) {
	// Mock LLM Server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := struct {
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
						Content: `{"command": "light on", "args": "", "thought": "resolved by LLM"}`,
					},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	config := Config{
		LLM: llm.Config{
			Endpoint: server.URL,
			Model:    "test-model",
		},
	}

	cmds := Commands{
		Definitions: []Definition{
			{Name: "light on", Description: "turn on the light", Factory: NewDummySubcommand},
			{Name: "aircon on", Description: "turn on the air conditioner", Factory: NewDummySubcommand},
		},
	}

	t.Run("LLM resolution success", func(t *testing.T) {
		def, args, msg, err := cmds.Find(t.Context(), config, "あかりをつけて")
		if err != nil {
			t.Fatalf("Find failed: %v", err)
		}
		if def.Name != "light on" {
			t.Errorf("expected command 'light on', got '%s'", def.Name)
		}
		if args != "" {
			t.Errorf("expected empty args, got '%s'", args)
		}
		if msg != "(LLM) resolved by LLM" {
			t.Errorf("expected LLM thought in msg, got '%s'", msg)
		}
	})

	t.Run("LLM resolution unknown command", func(t *testing.T) {
		// Override server response for unknown command
		server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprint(w, `{"choices":[{"message":{"content":"{\"command\": \"unknown\", \"args\": \"\", \"thought\": \"unknown command\"}"}}]}`)
		})

		_, _, _, err := cmds.Find(t.Context(), config, "何か未知の操作")
		if err == nil {
			t.Fatal("expected error for unknown command, got nil")
		}
	})

	t.Run("LLM resolution empty command", func(t *testing.T) {
		server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprint(w, `{"choices":[{"message":{"content":"{\"command\": \"\", \"args\": \"\", \"thought\": \"no match\"}"}}]}`)
		})

		_, _, _, err := cmds.Find(t.Context(), config, "何もしないで")
		if err == nil {
			t.Fatal("expected error for empty command, got nil")
		}
	})
}
