package subcommand

import (
	"encoding/json"
	"fmt"
	"github.com/johtani/smarthome/server/cron/influxdb"
	"github.com/johtani/smarthome/subcommand/action/owntone"
	"github.com/johtani/smarthome/subcommand/action/switchbot"
	"github.com/johtani/smarthome/subcommand/action/yamaha"
	"os"
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

func (c Config) validate() error {
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

func LoadConfig() (Config, error) {
	return LoadConfigWithPath(ConfigFileName)
}

func LoadConfigWithPath(configFile string) (Config, error) {
	file, err := os.Open(configFile)
	if err != nil {
		return Config{}, fmt.Errorf("設定ファイルの読み込みに失敗しました (%s): %w", configFile, err)
	}
	defer file.Close()

	// JSONデコード
	var config Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return Config{}, fmt.Errorf("設定ファイルのJSON解析に失敗しました: %w", err)
	}

	if err := config.validate(); err != nil {
		return Config{}, fmt.Errorf("設定のバリデーションに失敗しました:\n%w", err)
	}

	config.Commands = NewCommands()
	return config, nil
}
