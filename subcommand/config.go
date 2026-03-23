package subcommand

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/johtani/smarthome/server/cron/influxdb"
	"github.com/johtani/smarthome/subcommand/action/llm"
	"github.com/johtani/smarthome/subcommand/action/owntone"
	"github.com/johtani/smarthome/subcommand/action/switchbot"
	"github.com/johtani/smarthome/subcommand/action/yamaha"
	"os"
	"strconv"
	"strings"
)

// Config represents the application configuration.
type Config struct {
	Owntone   owntone.Config   `json:"Owntone"`
	Switchbot switchbot.Config `json:"Switchbot"`
	Yamaha    yamaha.Config    `json:"Yamaha"`
	LLM       llm.Config       `json:"LLM"`
	Influxdb  influxdb.Config  `json:"Influxdb"`
	Commands  Commands
}

// ConfigFileName is the default path to the configuration file.
const ConfigFileName = "./config/config.json"

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

// LoadConfig loads the configuration from the default path.
func LoadConfig() (Config, error) {
	return LoadConfigWithPath(ConfigFileName)
}

// LoadConfigWithPath loads the configuration from the specified path.
func LoadConfigWithPath(configFile string) (Config, error) {
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

	config.overrideWithEnv()

	if err := config.validate(); err != nil {
		return Config{}, fmt.Errorf("設定のバリデーションに失敗しました:\n%w", err)
	}

	config.Commands = NewCommands()
	return config, nil
}
