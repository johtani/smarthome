package subcommand

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/johtani/smarthome/server/cron/influxdb"
	"github.com/johtani/smarthome/subcommand/action/llm"
	"github.com/johtani/smarthome/subcommand/action/owntone"
	"github.com/johtani/smarthome/subcommand/action/switchbot"
	"github.com/johtani/smarthome/subcommand/action/yamaha"
)

// Config represents the application configuration.
type Config struct {
	Owntone   owntone.Config   `json:"Owntone"`
	Switchbot switchbot.Config `json:"Switchbot"`
	Yamaha    yamaha.Config    `json:"Yamaha"`
	LLM       llm.Config       `json:"LLM"`
	Resolver  ResolverConfig   `json:"Resolver"`
	Influxdb  influxdb.Config  `json:"Influxdb"`
	Commands  Commands
}

// ResolverConfig controls resolver behavior and observability options.
type ResolverConfig struct {
	Mode               string `json:"mode"`
	FeedbackEnabled    bool   `json:"feedback_enabled"`
	PromptVersion      string `json:"prompt_version"`
	DSPyEndpoint       string `json:"dspy_endpoint"`
	DSPyTimeoutSeconds int    `json:"dspy_timeout_seconds"`
}

const (
	// ResolverModeLegacy uses the current built-in resolution flow.
	ResolverModeLegacy = "legacy"
	// ResolverModeDSPy uses an external DSPy-based resolver flow.
	ResolverModeDSPy = "dspy"
)

func (c *ResolverConfig) applyDefaults() {
	if strings.TrimSpace(c.Mode) == "" {
		c.Mode = ResolverModeLegacy
	}
	if c.DSPyTimeoutSeconds <= 0 {
		c.DSPyTimeoutSeconds = 5
	}
}

// Validate validates ResolverConfig.
func (c ResolverConfig) Validate() error {
	switch c.Mode {
	case "", ResolverModeLegacy, ResolverModeDSPy:
		if c.DSPyTimeoutSeconds < 0 {
			return fmt.Errorf("resolver.dspy_timeout_seconds must be >= 0")
		}
		return nil
	default:
		return fmt.Errorf("resolver.mode must be one of %q or %q", ResolverModeLegacy, ResolverModeDSPy)
	}
}

// ConfigFileName is the default path to the configuration file.
// For backward compatibility with the -config flag.
const ConfigFileName = "./config/config.json"

// ConfigDirName is the default path to the configuration directory.
const ConfigDirName = "./config"

var knownConfigFiles = map[string]struct{}{
	"config.json":        {},
	"config.json.sample": {},
	"macros.json":        {},
	"slack.json":         {},
	"slack.json.sample":  {},
}

func (c *Config) validate() error {
	var errs []string
	var err error
	err = c.Owntone.Validate()
	if err != nil {
		errs = append(errs, err.Error())
	}
	err = c.Switchbot.Validate()
	if err != nil {
		errs = append(errs, err.Error())
	}
	err = c.Yamaha.Validate()
	if err != nil {
		errs = append(errs, err.Error())
	}
	err = c.LLM.Validate()
	if err != nil {
		errs = append(errs, err.Error())
	}
	err = c.Resolver.Validate()
	if err != nil {
		errs = append(errs, err.Error())
	}

	if len(errs) > 0 {
		return fmt.Errorf("%s", strings.Join(errs, "\n"))
	}
	return nil
}

func (c *Config) overrideWithEnv() {
	// SMARTHOME_OWNTONE_URL
	if val, ok := os.LookupEnv("SMARTHOME_OWNTONE_URL"); ok {
		c.Owntone.URL = val
	}
	// SMARTHOME_OWNTONE_TIMEOUT
	if val, ok := os.LookupEnv("SMARTHOME_OWNTONE_TIMEOUT"); ok {
		if i, err := strconv.Atoi(val); err == nil {
			c.Owntone.Timeout = i
		}
	}
	// SMARTHOME_OWNTONE_MUSIC_INTENT_ENDPOINT
	if val, ok := os.LookupEnv("SMARTHOME_OWNTONE_MUSIC_INTENT_ENDPOINT"); ok {
		c.Owntone.MusicIntentEndpoint = val
	}
	// SMARTHOME_OWNTONE_MUSIC_INTENT_TIMEOUT_SECONDS
	if val, ok := os.LookupEnv("SMARTHOME_OWNTONE_MUSIC_INTENT_TIMEOUT_SECONDS"); ok {
		if i, err := strconv.Atoi(val); err == nil {
			c.Owntone.MusicIntentTimeoutSeconds = i
		}
	}
	// SMARTHOME_OWNTONE_MUSIC_INTENT_CONFIDENCE_THRESHOLD
	if val, ok := os.LookupEnv("SMARTHOME_OWNTONE_MUSIC_INTENT_CONFIDENCE_THRESHOLD"); ok {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			c.Owntone.MusicIntentConfidenceThreshold = f
			c.Owntone.MusicIntentConfidenceThresholdSet = true
		}
	}
	// SMARTHOME_SWITCHBOT_TOKEN
	if val, ok := os.LookupEnv("SMARTHOME_SWITCHBOT_TOKEN"); ok {
		c.Switchbot.Token = val
	}
	// SMARTHOME_SWITCHBOT_SECRET
	if val, ok := os.LookupEnv("SMARTHOME_SWITCHBOT_SECRET"); ok {
		c.Switchbot.Secret = val
	}
	// SMARTHOME_YAMAHA_URL
	if val, ok := os.LookupEnv("SMARTHOME_YAMAHA_URL"); ok {
		c.Yamaha.URL = val
	}
	// SMARTHOME_YAMAHA_TIMEOUT
	if val, ok := os.LookupEnv("SMARTHOME_YAMAHA_TIMEOUT"); ok {
		if i, err := strconv.Atoi(val); err == nil {
			c.Yamaha.Timeout = i
		}
	}
	// SMARTHOME_LLM_API_KEY
	if val, ok := os.LookupEnv("SMARTHOME_LLM_API_KEY"); ok {
		c.LLM.APIKey = val
	}
	// SMARTHOME_LLM_ENDPOINT
	if val, ok := os.LookupEnv("SMARTHOME_LLM_ENDPOINT"); ok {
		c.LLM.Endpoint = val
	}
	// SMARTHOME_LLM_MODEL
	if val, ok := os.LookupEnv("SMARTHOME_LLM_MODEL"); ok {
		c.LLM.Model = val
	}
	// SMARTHOME_RESOLVER_MODE
	if val, ok := os.LookupEnv("SMARTHOME_RESOLVER_MODE"); ok {
		c.Resolver.Mode = val
	}
	// SMARTHOME_RESOLVER_FEEDBACK_ENABLED
	if val, ok := os.LookupEnv("SMARTHOME_RESOLVER_FEEDBACK_ENABLED"); ok {
		if b, err := strconv.ParseBool(val); err == nil {
			c.Resolver.FeedbackEnabled = b
		}
	}
	// SMARTHOME_RESOLVER_PROMPT_VERSION
	if val, ok := os.LookupEnv("SMARTHOME_RESOLVER_PROMPT_VERSION"); ok {
		c.Resolver.PromptVersion = val
	}
	// SMARTHOME_RESOLVER_DSPY_ENDPOINT
	if val, ok := os.LookupEnv("SMARTHOME_RESOLVER_DSPY_ENDPOINT"); ok {
		c.Resolver.DSPyEndpoint = val
	}
	// SMARTHOME_RESOLVER_DSPY_TIMEOUT_SECONDS
	if val, ok := os.LookupEnv("SMARTHOME_RESOLVER_DSPY_TIMEOUT_SECONDS"); ok {
		if i, err := strconv.Atoi(val); err == nil {
			c.Resolver.DSPyTimeoutSeconds = i
		}
	}
	// SMARTHOME_INFLUXDB_TOKEN
	if val, ok := os.LookupEnv("SMARTHOME_INFLUXDB_TOKEN"); ok {
		c.Influxdb.Token = val
	}
	// SMARTHOME_INFLUXDB_URL
	if val, ok := os.LookupEnv("SMARTHOME_INFLUXDB_URL"); ok {
		c.Influxdb.URL = val
	}
	// SMARTHOME_INFLUXDB_BUCKET
	if val, ok := os.LookupEnv("SMARTHOME_INFLUXDB_BUCKET"); ok {
		c.Influxdb.Bucket = val
	}
	// SMARTHOME_INFLUXDB_ORG
	if val, ok := os.LookupEnv("SMARTHOME_INFLUXDB_ORG"); ok {
		c.Influxdb.Org = val
	}
	// SMARTHOME_INFLUXDB_MEASUREMENT
	if val, ok := os.LookupEnv("SMARTHOME_INFLUXDB_MEASUREMENT"); ok {
		c.Influxdb.Measurement = val
	}
}

func detectOwntoneThresholdSet(configFile string) (bool, error) {
	// #nosec G304
	data, err := os.ReadFile(configFile)
	if err != nil {
		return false, err
	}

	var root map[string]json.RawMessage
	if err := json.Unmarshal(data, &root); err != nil {
		return false, nil
	}

	var owntoneRaw json.RawMessage
	for _, key := range []string{"owntone", "Owntone"} {
		if v, ok := root[key]; ok {
			owntoneRaw = v
			break
		}
	}
	if len(owntoneRaw) == 0 {
		return false, nil
	}

	var owntoneMap map[string]json.RawMessage
	if err := json.Unmarshal(owntoneRaw, &owntoneMap); err != nil {
		return false, nil
	}

	_, ok := owntoneMap["music_intent_confidence_threshold"]
	return ok, nil
}

// LoadConfig loads the configuration from the default directory.
func LoadConfig() (Config, error) {
	return LoadConfigFromDir(ConfigDirName)
}

// LoadConfigFromDir loads configuration from a directory.
// It reads config.json (required) and macros.json (optional) from the directory.
// Unknown files in the directory are logged as warnings.
func LoadConfigFromDir(dir string) (Config, error) {
	config, err := loadConfigJSON(filepath.Join(dir, "config.json"))
	if err != nil {
		return Config{}, err
	}

	macros, err := loadMacrosFromFile(filepath.Join(dir, "macros.json"))
	if err != nil {
		return Config{}, err
	}

	warnUnknownFiles(dir)

	config.Commands = NewCommands(macros...)
	return config, nil
}

// LoadConfigWithPath loads the configuration from the specified file path.
// For backward compatibility with the -config flag.
func LoadConfigWithPath(configFile string) (Config, error) {
	config, err := loadConfigJSON(configFile)
	if err != nil {
		return Config{}, err
	}
	config.Commands = NewCommands()
	return config, nil
}

func loadConfigJSON(configFile string) (Config, error) {
	// #nosec G304
	file, err := os.Open(configFile)
	if err != nil {
		return Config{}, fmt.Errorf("設定ファイルの読み込みに失敗しました (%s): %w", configFile, err)
	}
	defer func() {
		_ = file.Close()
	}()

	// JSONデコード
	var config Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		var typeErr *json.UnmarshalTypeError
		var syntaxErr *json.SyntaxError
		if errors.As(err, &typeErr) {
			return Config{}, fmt.Errorf("invalid value for field '%s': expected %s, got %s", typeErr.Field, typeErr.Type, typeErr.Value)
		} else if errors.As(err, &syntaxErr) {
			return Config{}, fmt.Errorf("JSON syntax error at byte offset %d: %w", syntaxErr.Offset, syntaxErr)
		}
		return Config{}, fmt.Errorf("failed to parse config file: %w", err)
	}

	thresholdSet, detectErr := detectOwntoneThresholdSet(configFile)
	if detectErr == nil {
		config.Owntone.MusicIntentConfidenceThresholdSet = thresholdSet
	}

	config.overrideWithEnv()
	config.Resolver.applyDefaults()

	if err := config.validate(); err != nil {
		return Config{}, fmt.Errorf("設定のバリデーションに失敗しました:\n%w", err)
	}

	return config, nil
}

func loadMacrosFromFile(path string) ([]MacroConfig, error) {
	// #nosec G304
	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to open macros file (%s): %w", path, err)
	}
	defer func() {
		_ = file.Close()
	}()

	var macros []MacroConfig
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&macros); err != nil {
		var typeErr *json.UnmarshalTypeError
		var syntaxErr *json.SyntaxError
		if errors.As(err, &typeErr) {
			return nil, fmt.Errorf("invalid value for field '%s' in macros: expected %s, got %s", typeErr.Field, typeErr.Type, typeErr.Value)
		} else if errors.As(err, &syntaxErr) {
			return nil, fmt.Errorf("JSON syntax error in macros file at byte offset %d: %w", syntaxErr.Offset, syntaxErr)
		}
		return nil, fmt.Errorf("failed to parse macros file: %w", err)
	}

	if err := validateMacros(macros); err != nil {
		return nil, err
	}

	return macros, nil
}

func warnUnknownFiles(dir string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if _, ok := knownConfigFiles[entry.Name()]; !ok {
			slog.Warn("unknown file in config directory, ignored", "file", entry.Name(), "dir", dir)
		}
	}
}
