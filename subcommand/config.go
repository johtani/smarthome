package subcommand

import (
	"encoding/json"
	"fmt"
	"github.com/johtani/smarthome/server/cron/influxdb"
	"github.com/johtani/smarthome/subcommand/action/owntone"
	"github.com/johtani/smarthome/subcommand/action/switchbot"
	"github.com/johtani/smarthome/subcommand/action/yamaha"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Owntone   owntone.Config   `json:"Owntone"`
	Switchbot switchbot.Config `json:"Switchbot"`
	Yamaha    yamaha.Config    `json:"Yamaha"`
	Influxdb  influxdb.Config  `json:"Influxdb"`
	Commands  Commands
}

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

func LoadConfig() (Config, error) {
	return LoadConfigWithPath(ConfigFileName)
}

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
		return Config{}, fmt.Errorf("設定ファイルのJSON解析に失敗しました: %w", err)
	}

	config.overrideWithEnv()

	if err := config.validate(); err != nil {
		return Config{}, fmt.Errorf("設定のバリデーションに失敗しました:\n%w", err)
	}

	config.Commands = NewCommands()
	return config, nil
}
