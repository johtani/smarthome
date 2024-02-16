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
			newEntry(StartMeetingCmd, NewStartMeetingDefinition(), []string{}),
			newEntry(FinishMeetingCmd, NewFinishMeetingDefinition(), []string{}),
			newEntry(StartMusicCmd, NewStartMusicCmdDefinition(), []string{}),
			newEntry(StopMusicCmd, NewStopMusicDefinition(), []string{}),
			newEntry(ChangePlaylistCmd, NewChangePlaylistCmdDefinition(), []string{}),
			newEntry(SwitchBotDeviceListCmd, NewSwitchBotDeviceListDefinition(), []string{}),
			newEntry(DeviceListCmd, NewSwitchBotDeviceListDefinition(), []string{}),
			newEntry(SwitchBotSceneListCmd, NewSwitchBotSceneListDefinition(), []string{}),
			newEntry(SceneListCmd, NewSwitchBotSceneListDefinition(), []string{}),
			newEntry(LightOffCmd, NewLightOffDefinition(), []string{}),
			newEntry(LightOnCmd, NewLightOnDefinition(), []string{}),
			newEntry(HelpCmd, NewHelpDefinition(), []string{}),
			newEntry(StartSwitchCmd, NewStartSwitchDefinition(), []string{}),
			newEntry(StartPS5Cmd, NewStartPS5Definition(), []string{}),
			newEntry(AirConditionerOnCmd, NewAirConditionerOnDefinition(), []string{ACOnCmd}),
			newEntry(AirConditionerOffCmd, NewAirConditionerOffDefinition(), []string{ACOffCmd}),
			newEntry(DisplayTemperatureCmd, NewDisplayTemperatureDefinition(), []string{DispTempCmd}),
			newEntry(TokenizeIpaCmd, NewTokenizeIpaDefinition(), []string{}),
			newEntry(TokenizeUniCmd, NewTokenizeUniDefinition(), []string{}),
			newEntry(TokenizeNeologdCmd, NewTokenizeNeologdDefinition(), []string{}),
		},
	}
	return config
}
