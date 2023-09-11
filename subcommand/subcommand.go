package subcommand

import (
	"encoding/json"
	"fmt"
	"github.com/johtani/smarthome/subcommand/action"
	"github.com/johtani/smarthome/subcommand/action/owntone"
	"github.com/johtani/smarthome/subcommand/action/switchbot"
	"github.com/johtani/smarthome/subcommand/action/yamaha"
	"os"
	"strings"
)

type Subcommand struct {
	Definition
	actions     []action.Action
	ignoreError bool
}

type Definition struct {
	Name        string
	Description string
	Factory     func(Definition, Config) Subcommand
}

func (s Subcommand) Exec() (string, error) {
	var msgs []string
	for i := range s.actions {
		msg, err := s.actions[i].Run()
		if s.ignoreError && err != nil {
			fmt.Printf("skip error\t %v\n", err)
			//TODO msgsにエラーを追加する？
		} else if err != nil {
			return "", err
		}
		msgs = append(msgs, msg)
	}
	return strings.Join(msgs, "\n"), nil
}

func (d Definition) Init(config Config) Subcommand {
	return d.Factory(d, config)
}

func Map() map[string]Definition {
	return map[string]Definition{
		StartMeetingCmd:        NewStartMeetingDefinition(),
		FinishMeetingCmd:       NewFinishMeetingDefinition(),
		StartMusicCmd:          NewStartMusicCmdDefinition(),
		StopMusicCmd:           NewStopMusicDefinition(),
		SwitchBotDeviceListCmd: NewSwitchBotDeviceListDefinition(),
		DeviceListCmd:          NewSwitchBotDeviceListDefinition(),
		SwitchBotSceneListCmd:  NewSwitchBotSceneListDefinition(),
		SceneListCmd:           NewSwitchBotSceneListDefinition(),
		LightOffCmd:            NewLightOffDefinition(),
		LightOnCmd:             NewLightOnDefinition(),
		HelpCmd:                NewHelpDefinition(),
		StartSwitchCmd:         NewStartSwitchDefinition(),
		StartPS5Cmd:            NewStartPS5Definition(),
		AirConditionerOnCmd:    NewAirConditionerOnDefinition(),
		ACOnCmd:                NewAirConditionerOnDefinition(),
		AirConditionerOffCmd:   NewAirConditionerOffDefinition(),
		ACOffCmd:               NewAirConditionerOffDefinition(),
		DisplayTemperature:     NewDisplayTemperatureDefinition(),
		DispTemp:               NewDisplayTemperatureDefinition(),
	}
}

type Config struct {
	Owntone   owntone.Config   `json:"Owntone"`
	Switchbot switchbot.Config `json:"Switchbot"`
	Yamaha    yamaha.Config    `json:"Yamaha"`
}

const ConfigFileName = "./config/config.json"

func LoadConfig() Config {
	file, err := os.Open(ConfigFileName)
	if err != nil {
		panic(fmt.Sprintf("ファイルの読み込みエラー: %v", err))
	}
	// JSONデコード
	decoder := json.NewDecoder(file)
	var config Config
	err = decoder.Decode(&config)
	if err != nil {
		panic(fmt.Sprintf("JSONデコードエラー: %v", err))
	}
	err = config.validate()
	if err != nil {
		panic(fmt.Sprintf("Validation エラー: \n%v", err))
	}
	return config
}

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
		return fmt.Errorf(strings.Join(errs, "\n"))
	}
	return nil
}
