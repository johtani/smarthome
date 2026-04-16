package subcommand

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
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

func TestCommands_Find_DSPyMode(t *testing.T) {
	t.Run("dspy success", func(t *testing.T) {
		dspyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprint(w, `{"command":"light on","args":"","thought":"resolved by DSPy"}`)
		}))
		defer dspyServer.Close()

		config := Config{
			Resolver: ResolverConfig{
				Mode:               ResolverModeDSPy,
				DSPyEndpoint:       dspyServer.URL,
				DSPyTimeoutSeconds: 3,
			},
			LLM: llm.Config{
				Endpoint: "http://unused",
				Model:    "test-model",
			},
		}

		cmds := Commands{
			Definitions: []Definition{
				{Name: "light on", Description: "turn on the light", Factory: NewDummySubcommand},
			},
		}

		def, args, msg, err := cmds.Find(t.Context(), config, "あかりをつけて")
		if err != nil {
			t.Fatalf("Find failed: %v", err)
		}
		if def.Name != "light on" {
			t.Fatalf("expected command 'light on', got %q", def.Name)
		}
		if args != "" {
			t.Fatalf("expected empty args, got %q", args)
		}
		if msg != "(DSPy) resolved by DSPy" {
			t.Fatalf("unexpected msg: %q", msg)
		}
	})

	t.Run("dspy failure falls back to llm", func(t *testing.T) {
		var llmCalled atomic.Int32
		dspyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "dspy unavailable", http.StatusServiceUnavailable)
		}))
		defer dspyServer.Close()

		llmServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			llmCalled.Add(1)
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprint(w, `{"choices":[{"message":{"content":"{\"command\": \"light on\", \"args\": \"\", \"thought\": \"resolved by LLM\"}"}}]}`)
		}))
		defer llmServer.Close()

		config := Config{
			Resolver: ResolverConfig{
				Mode:               ResolverModeDSPy,
				DSPyEndpoint:       dspyServer.URL,
				DSPyTimeoutSeconds: 3,
			},
			LLM: llm.Config{
				Endpoint: llmServer.URL,
				Model:    "test-model",
			},
		}

		cmds := Commands{
			Definitions: []Definition{
				{Name: "light on", Description: "turn on the light", Factory: NewDummySubcommand},
			},
		}

		def, _, msg, err := cmds.Find(t.Context(), config, "あかりをつけて")
		if err != nil {
			t.Fatalf("Find failed: %v", err)
		}
		if def.Name != "light on" {
			t.Fatalf("expected command 'light on', got %q", def.Name)
		}
		if llmCalled.Load() != 1 {
			t.Fatalf("expected llm fallback to be called once, got %d", llmCalled.Load())
		}
		if msg != "(LLM) resolved by LLM" {
			t.Fatalf("unexpected msg: %q", msg)
		}
	})
}

func TestCommands_Find_LLM_FallbackStartMusicArgsToSearchAndPlay(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"choices":[{"message":{"content":"{\"command\": \"start music\", \"args\": \"Meja\", \"thought\": \"music request\"}"}}]}`)
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
			{Name: StartMusicCmd, Description: "legacy random music", Factory: NewDummySubcommand},
			{Name: SearchAndPlayMusicCmd, Description: "search and play music", Factory: NewDummySubcommand},
		},
	}

	def, args, msg, err := cmds.Find(t.Context(), config, "Mejaを再生して")
	if err != nil {
		t.Fatalf("Find failed: %v", err)
	}
	if def.Name != SearchAndPlayMusicCmd {
		t.Fatalf("expected command %q, got %q", SearchAndPlayMusicCmd, def.Name)
	}
	if args != "Meja" {
		t.Fatalf("expected args %q, got %q", "Meja", args)
	}
	if msg != "(LLM) music request" {
		t.Fatalf("unexpected msg: %s", msg)
	}
}
