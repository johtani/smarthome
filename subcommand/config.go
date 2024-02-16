package subcommand

import (
	"encoding/json"
	"fmt"
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
		return fmt.Errorf(strings.Join(errs, "\n"))
	}
	return nil
}

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
	config.Commands = Commands{
		entries: []Entry{
			newEntry(StartMeetingCmd, NewStartMeetingDefinition()),
			newEntry(FinishMeetingCmd, NewFinishMeetingDefinition()),
			newEntry(StartMusicCmd, NewStartMusicCmdDefinition()),
			newEntry(StopMusicCmd, NewStopMusicDefinition()),
			newEntry(ChangePlaylistCmd, NewChangePlaylistCmdDefinition()),
			newEntry(SwitchBotDeviceListCmd, NewSwitchBotDeviceListDefinition()),
			newEntry(DeviceListCmd, NewSwitchBotDeviceListDefinition()),
			newEntry(SwitchBotSceneListCmd, NewSwitchBotSceneListDefinition()),
			newEntry(SceneListCmd, NewSwitchBotSceneListDefinition()),
			newEntry(LightOffCmd, NewLightOffDefinition()),
			newEntry(LightOnCmd, NewLightOnDefinition()),
			newEntry(HelpCmd, NewHelpDefinition()),
			newEntry(StartSwitchCmd, NewStartSwitchDefinition()),
			newEntry(StartPS5Cmd, NewStartPS5Definition()),
			newEntry(AirConditionerOnCmd, NewAirConditionerOnDefinition()),
			newEntry(AirConditionerOffCmd, NewAirConditionerOffDefinition()),
			newEntry(DisplayTemperatureCmd, NewDisplayTemperatureDefinition()),
			newEntry(TokenizeIpaCmd, NewTokenizeIpaDefinition()),
			newEntry(TokenizeUniCmd, NewTokenizeUniDefinition()),
			newEntry(TokenizeNeologdCmd, NewTokenizeNeologdDefinition()),
		},
	}
	return config
}
