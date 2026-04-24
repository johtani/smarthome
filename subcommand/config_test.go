package subcommand

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/johtani/smarthome/subcommand/action/llm"
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
				Owntone:   owntone.Config{URL: "http://localhost:8000"},
				Switchbot: switchbot.Config{Token: "token", Secret: "secret"},
				Yamaha:    yamaha.Config{URL: "http://localhost:8080"},
				LLM:       llm.Config{Endpoint: "http://localhost:8081", Model: "gpt-4o"},
				Resolver:  ResolverConfig{Mode: ResolverModeLegacy},
			},
			wantErr: false,
		},
		{
			name: "invalid owntone config",
			config: Config{
				Owntone:   owntone.Config{URL: ""},
				Switchbot: switchbot.Config{Token: "token", Secret: "secret"},
				Yamaha:    yamaha.Config{URL: "http://localhost:8080"},
				LLM:       llm.Config{Endpoint: "http://localhost:8081", Model: "gpt-4o"},
				Resolver:  ResolverConfig{Mode: ResolverModeLegacy},
			},
			wantErr: true,
		},
		{
			name: "invalid llm config",
			config: Config{
				Owntone:   owntone.Config{URL: "http://localhost:8000"},
				Switchbot: switchbot.Config{Token: "token", Secret: "secret"},
				Yamaha:    yamaha.Config{URL: "http://localhost:8080"},
				LLM:       llm.Config{Endpoint: "http://localhost:8081"},
				Resolver:  ResolverConfig{Mode: ResolverModeLegacy},
			},
			wantErr: true,
		},
		{
			name: "invalid resolver mode",
			config: Config{
				Owntone:   owntone.Config{URL: "http://localhost:8000"},
				Switchbot: switchbot.Config{Token: "token", Secret: "secret"},
				Yamaha:    yamaha.Config{URL: "http://localhost:8080"},
				LLM:       llm.Config{Endpoint: "http://localhost:8081", Model: "gpt-4o"},
				Resolver:  ResolverConfig{Mode: "unknown"},
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
	_ = os.Setenv("SMARTHOME_OWNTONE_URL", "http://env-owntone")
	_ = os.Setenv("SMARTHOME_SWITCHBOT_TOKEN", "env-token")
	_ = os.Setenv("SMARTHOME_SWITCHBOT_SECRET", "env-secret")
	_ = os.Setenv("SMARTHOME_YAMAHA_URL", "http://env-yamaha")
	_ = os.Setenv("SMARTHOME_LLM_API_KEY", "env-llm-key")
	_ = os.Setenv("SMARTHOME_LLM_ENDPOINT", "http://env-llm-endpoint")
	_ = os.Setenv("SMARTHOME_LLM_MODEL", "env-llm-model")
	_ = os.Setenv("SMARTHOME_RESOLVER_MODE", ResolverModeDSPy)
	_ = os.Setenv("SMARTHOME_RESOLVER_FEEDBACK_ENABLED", "true")
	_ = os.Setenv("SMARTHOME_RESOLVER_PROMPT_VERSION", "2026-04-14")
	_ = os.Setenv("SMARTHOME_RESOLVER_DSPY_ENDPOINT", "http://env-dspy-endpoint")
	_ = os.Setenv("SMARTHOME_RESOLVER_DSPY_TIMEOUT_SECONDS", "9")
	_ = os.Setenv("SMARTHOME_OWNTONE_MUSIC_INTENT_CONFIDENCE_THRESHOLD", "0")
	_ = os.Setenv("SMARTHOME_INFLUXDB_TOKEN", "env-influx-token")
	_ = os.Setenv("SMARTHOME_INFLUXDB_URL", "http://env-influx-url")
	_ = os.Setenv("SMARTHOME_INFLUXDB_BUCKET", "env-bucket")
	defer func() {
		_ = os.Unsetenv("SMARTHOME_OWNTONE_URL")
		_ = os.Unsetenv("SMARTHOME_SWITCHBOT_TOKEN")
		_ = os.Unsetenv("SMARTHOME_SWITCHBOT_SECRET")
		_ = os.Unsetenv("SMARTHOME_YAMAHA_URL")
		_ = os.Unsetenv("SMARTHOME_LLM_API_KEY")
		_ = os.Unsetenv("SMARTHOME_LLM_ENDPOINT")
		_ = os.Unsetenv("SMARTHOME_LLM_MODEL")
		_ = os.Unsetenv("SMARTHOME_RESOLVER_MODE")
		_ = os.Unsetenv("SMARTHOME_RESOLVER_FEEDBACK_ENABLED")
		_ = os.Unsetenv("SMARTHOME_RESOLVER_PROMPT_VERSION")
		_ = os.Unsetenv("SMARTHOME_RESOLVER_DSPY_ENDPOINT")
		_ = os.Unsetenv("SMARTHOME_RESOLVER_DSPY_TIMEOUT_SECONDS")
		_ = os.Unsetenv("SMARTHOME_OWNTONE_MUSIC_INTENT_CONFIDENCE_THRESHOLD")
		_ = os.Unsetenv("SMARTHOME_INFLUXDB_TOKEN")
		_ = os.Unsetenv("SMARTHOME_INFLUXDB_URL")
		_ = os.Unsetenv("SMARTHOME_INFLUXDB_BUCKET")
	}()

	config.overrideWithEnv()

	if config.Owntone.URL != "http://env-owntone" {
		t.Errorf("expected http://env-owntone, got %s", config.Owntone.URL)
	}
	if config.Switchbot.Token != "env-token" {
		t.Errorf("expected env-token, got %s", config.Switchbot.Token)
	}
	if config.Switchbot.Secret != "env-secret" {
		t.Errorf("expected env-secret, got %s", config.Switchbot.Secret)
	}
	if config.Yamaha.URL != "http://env-yamaha" {
		t.Errorf("expected http://env-yamaha, got %s", config.Yamaha.URL)
	}
	if config.LLM.APIKey != "env-llm-key" {
		t.Errorf("expected env-llm-key, got %s", config.LLM.APIKey)
	}
	if config.LLM.Endpoint != "http://env-llm-endpoint" {
		t.Errorf("expected http://env-llm-endpoint, got %s", config.LLM.Endpoint)
	}
	if config.LLM.Model != "env-llm-model" {
		t.Errorf("expected env-llm-model, got %s", config.LLM.Model)
	}
	if config.Resolver.Mode != ResolverModeDSPy {
		t.Errorf("expected %s, got %s", ResolverModeDSPy, config.Resolver.Mode)
	}
	if !config.Resolver.FeedbackEnabled {
		t.Error("expected feedback enabled to be true")
	}
	if config.Resolver.PromptVersion != "2026-04-14" {
		t.Errorf("expected prompt version 2026-04-14, got %s", config.Resolver.PromptVersion)
	}
	if config.Resolver.DSPyEndpoint != "http://env-dspy-endpoint" {
		t.Errorf("expected dspy endpoint http://env-dspy-endpoint, got %s", config.Resolver.DSPyEndpoint)
	}
	if config.Resolver.DSPyTimeoutSeconds != 9 {
		t.Errorf("expected dspy timeout 9, got %d", config.Resolver.DSPyTimeoutSeconds)
	}
	if config.Owntone.MusicIntentConfidenceThreshold != 0 {
		t.Errorf("expected owntone threshold 0, got %v", config.Owntone.MusicIntentConfidenceThreshold)
	}
	if !config.Owntone.MusicIntentConfidenceThresholdSet {
		t.Error("expected owntone threshold set flag to be true")
	}
	if config.Influxdb.Token != "env-influx-token" {
		t.Errorf("expected env-influx-token, got %s", config.Influxdb.Token)
	}
	if config.Influxdb.URL != "http://env-influx-url" {
		t.Errorf("expected http://env-influx-url, got %s", config.Influxdb.URL)
	}
	if config.Influxdb.Bucket != "env-bucket" {
		t.Errorf("expected env-bucket, got %s", config.Influxdb.Bucket)
	}
}

func TestLoadConfigWithPath_JSONErrors(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		errContains string
	}{
		{
			name: "type mismatch",
			content: `{
				"Owntone": {"url": "http://localhost:8000", "timeout": "not-a-number"},
				"Switchbot": {"token": "token", "secret": "secret"},
				"Yamaha": {"url": "http://localhost:8080"},
				"LLM": {"endpoint": "http://localhost:8081", "model": "gpt-4o"}
			}`,
			errContains: "invalid value for field",
		},
		{
			name:        "syntax error",
			content:     `{"Owntone": {invalid}}`,
			errContains: "JSON syntax error at byte offset",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpfile, err := os.CreateTemp("", "config_test_*.json")
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				_ = os.Remove(tmpfile.Name())
			}()
			if _, err := tmpfile.Write([]byte(tt.content)); err != nil {
				t.Fatal(err)
			}
			if err := tmpfile.Close(); err != nil {
				t.Fatal(err)
			}

			_, err = LoadConfigWithPath(tmpfile.Name())
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("expected error to contain %q, got: %v", tt.errContains, err)
			}
		})
	}
}

func TestLoadConfigWithPath(t *testing.T) {
	// 実際の設定ファイルを使うのは避けたいので、一時的なファイルを作成する
	content := `{
		"Owntone": {"url": "http://localhost:8000"},
		"Switchbot": {"token": "token", "secret": "secret"},
		"Yamaha": {"url": "http://localhost:8080"},
		"LLM": {"endpoint": "http://localhost:8081", "model": "gpt-4o"},
		"Influxdb": {"url": "http://localhost:8086", "token": "token", "bucket": "bucket", "org": "org"}
	}`
	tmpfile, err := os.CreateTemp("", "config_test.json")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Remove(tmpfile.Name())
	}()

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

	if config.Owntone.URL != "http://localhost:8000" {
		t.Errorf("expected http://localhost:8000, got %s", config.Owntone.URL)
	}
	if len(config.Commands.Definitions) == 0 {
		t.Error("commands should be initialized")
	}
	if config.Resolver.Mode != ResolverModeLegacy {
		t.Errorf("expected resolver mode %s, got %s", ResolverModeLegacy, config.Resolver.Mode)
	}
	if config.Resolver.DSPyTimeoutSeconds != 5 {
		t.Errorf("expected default dspy timeout 5, got %d", config.Resolver.DSPyTimeoutSeconds)
	}
	if config.Owntone.MusicIntentConfidenceThresholdSet {
		t.Error("expected owntone threshold set flag to be false when key is missing")
	}
}

func TestLoadConfigWithPath_OwntoneThresholdExplicitlySet(t *testing.T) {
	content := `{
		"Owntone": {"url": "http://localhost:8000", "music_intent_confidence_threshold": 0},
		"Switchbot": {"token": "token", "secret": "secret"},
		"Yamaha": {"url": "http://localhost:8080"},
		"LLM": {"endpoint": "http://localhost:8081", "model": "gpt-4o"},
		"Influxdb": {"url": "http://localhost:8086", "token": "token", "bucket": "bucket", "org": "org"}
	}`
	tmpfile, err := os.CreateTemp("", "config_test_threshold.json")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Remove(tmpfile.Name())
	}()

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

	if !config.Owntone.MusicIntentConfidenceThresholdSet {
		t.Error("expected owntone threshold set flag to be true when key exists")
	}
	if config.Owntone.MusicIntentConfidenceThreshold != 0 {
		t.Errorf("expected threshold 0, got %v", config.Owntone.MusicIntentConfidenceThreshold)
	}
}

func TestLoadConfigWithPath_LLMDisabled(t *testing.T) {
	content := `{
		"Owntone": {"url": "http://localhost:8000"},
		"Switchbot": {"token": "token", "secret": "secret"},
		"Yamaha": {"url": "http://localhost:8080"},
		"LLM": {},
		"Influxdb": {"url": "http://localhost:8086", "token": "token", "bucket": "bucket", "org": "org"}
	}`
	tmpfile, err := os.CreateTemp("", "config_test_no_llm.json")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Remove(tmpfile.Name())
	}()

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

	if config.LLM.Endpoint != "" {
		t.Errorf("expected empty llm endpoint, got %s", config.LLM.Endpoint)
	}
	if config.LLM.Model != "" {
		t.Errorf("expected empty llm model, got %s", config.LLM.Model)
	}
}

const testConfigJSON = `{
	"Owntone": {"url": "http://localhost:8000"},
	"Switchbot": {"token": "token", "secret": "secret"},
	"Yamaha": {"url": "http://localhost:8080"},
	"LLM": {"endpoint": "http://localhost:8081", "model": "gpt-4o"}
}`

func TestLoadConfigFromDir(t *testing.T) {
	t.Run("config only (no macros.json)", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte(testConfigJSON), 0600); err != nil {
			t.Fatal(err)
		}

		config, err := LoadConfigFromDir(dir)
		if err != nil {
			t.Fatalf("LoadConfigFromDir failed: %v", err)
		}
		if config.Owntone.URL != "http://localhost:8000" {
			t.Errorf("expected http://localhost:8000, got %s", config.Owntone.URL)
		}
		if len(config.Commands.Definitions) == 0 {
			t.Error("commands should be initialized")
		}
	})

	t.Run("with valid macros.json", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte(testConfigJSON), 0600); err != nil {
			t.Fatal(err)
		}
		macrosJSON := `[{"name": "test macro", "description": "test", "ignore_error": true, "actions": [{"type": "owntone_pause"}]}]`
		if err := os.WriteFile(filepath.Join(dir, "macros.json"), []byte(macrosJSON), 0600); err != nil {
			t.Fatal(err)
		}

		config, err := LoadConfigFromDir(dir)
		if err != nil {
			t.Fatalf("LoadConfigFromDir failed: %v", err)
		}
		found := false
		for _, def := range config.Commands.Definitions {
			if def.Name == "test macro" {
				found = true
				break
			}
		}
		if !found {
			t.Error("macro 'test macro' should be registered as a command")
		}
	})

	t.Run("macro action build failure returns explicit runtime error", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte(testConfigJSON), 0600); err != nil {
			t.Fatal(err)
		}
		macrosJSON := `[{"name":"broken macro","description":"broken","ignore_error":true,"actions":[{"type":"wait","params":{"seconds":"abc"}}]}]`
		if err := os.WriteFile(filepath.Join(dir, "macros.json"), []byte(macrosJSON), 0600); err != nil {
			t.Fatal(err)
		}

		config, err := LoadConfigFromDir(dir)
		if err != nil {
			t.Fatalf("LoadConfigFromDir failed: %v", err)
		}

		def, args, _, err := config.Commands.Find(context.Background(), config, "broken macro")
		if err != nil {
			t.Fatalf("Find failed: %v", err)
		}

		_, err = def.Init(config).Exec(context.Background(), args)
		if err == nil {
			t.Fatal("expected execution error for broken macro")
		}
		if !strings.Contains(err.Error(), `macro "broken macro" initialization failed`) {
			t.Fatalf("expected macro initialization error, got %v", err)
		}
		if !strings.Contains(err.Error(), `invalid seconds "abc"`) {
			t.Fatalf("expected root cause in error, got %v", err)
		}
	})

	t.Run("config.json not found", func(t *testing.T) {
		dir := t.TempDir()
		_, err := LoadConfigFromDir(dir)
		if err == nil {
			t.Fatal("expected error when config.json is missing")
		}
	})

	t.Run("invalid macros.json", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte(testConfigJSON), 0600); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "macros.json"), []byte(`[{invalid}]`), 0600); err != nil {
			t.Fatal(err)
		}
		_, err := LoadConfigFromDir(dir)
		if err == nil {
			t.Fatal("expected error for invalid macros.json")
		}
	})
}

func TestLoadMacrosFromFile(t *testing.T) {
	t.Run("file not found returns nil", func(t *testing.T) {
		macros, err := loadMacrosFromFile("/nonexistent/path/macros.json")
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
		if macros != nil {
			t.Errorf("expected nil macros, got %v", macros)
		}
	})

	t.Run("valid macros file", func(t *testing.T) {
		content := `[
			{"name": "macro1", "description": "test", "ignore_error": true, "actions": [{"type": "owntone_pause"}]},
			{"name": "macro2", "description": "test2", "actions": [{"type": "yamaha_power_on"}]}
		]`
		f, err := os.CreateTemp("", "macros_test_*.json")
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Remove(f.Name()) }()
		if _, err := f.WriteString(content); err != nil {
			t.Fatal(err)
		}
		_ = f.Close()

		macros, err := loadMacrosFromFile(f.Name())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(macros) != 2 {
			t.Errorf("expected 2 macros, got %d", len(macros))
		}
		if macros[0].Name != "macro1" {
			t.Errorf("expected macro1, got %s", macros[0].Name)
		}
	})

	t.Run("JSON syntax error", func(t *testing.T) {
		f, err := os.CreateTemp("", "macros_test_*.json")
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Remove(f.Name()) }()
		if _, err := f.WriteString(`[{invalid}]`); err != nil {
			t.Fatal(err)
		}
		_ = f.Close()

		_, err = loadMacrosFromFile(f.Name())
		if err == nil {
			t.Fatal("expected error for invalid JSON")
		}
		if !strings.Contains(err.Error(), "JSON syntax error") {
			t.Errorf("expected JSON syntax error message, got: %v", err)
		}
	})

	t.Run("unknown action type", func(t *testing.T) {
		content := `[{"name": "bad", "actions": [{"type": "unknown_action"}]}]`
		f, err := os.CreateTemp("", "macros_test_*.json")
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Remove(f.Name()) }()
		if _, err := f.WriteString(content); err != nil {
			t.Fatal(err)
		}
		_ = f.Close()

		_, err = loadMacrosFromFile(f.Name())
		if err == nil {
			t.Fatal("expected validation error for unknown action type")
		}
		if !strings.Contains(err.Error(), "unknown action type") {
			t.Errorf("expected unknown action type error, got: %v", err)
		}
	})

	t.Run("empty actions", func(t *testing.T) {
		content := `[{"name": "bad", "actions": []}]`
		f, err := os.CreateTemp("", "macros_test_*.json")
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Remove(f.Name()) }()
		if _, err := f.WriteString(content); err != nil {
			t.Fatal(err)
		}
		_ = f.Close()

		_, err = loadMacrosFromFile(f.Name())
		if err == nil {
			t.Fatal("expected validation error for empty actions")
		}
		if !strings.Contains(err.Error(), "actions must not be empty") {
			t.Errorf("expected empty actions error, got: %v", err)
		}
	})
}
