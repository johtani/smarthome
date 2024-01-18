package subcommand

import (
	"encoding/json"
	"fmt"
	"github.com/hbollon/go-edlib"
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

type Entry struct {
	Name       string
	definition Definition
	shortnames []string
	noHyphens  []string
	Help       string
}

func newEntry(name string, definition Definition, shortnames []string) Entry {
	noHyphens := []string{strings.ReplaceAll(name, "-", " ")}
	for _, shortname := range shortnames {
		noHyphens = append(noHyphens, strings.ReplaceAll(shortname, "-", " "))
	}
	var help string
	if len(shortnames) > 0 {
		help = fmt.Sprintf("  %s [%s]: %s\n", name, strings.Join(shortnames, "/"), definition.Description)
	} else {
		help = fmt.Sprintf("  %s : %s\n", name, definition.Description)
	}
	return Entry{Name: name, definition: definition, shortnames: shortnames, noHyphens: noHyphens, Help: help}
}

func (e Entry) IsTarget(name string, withoutHyphen bool) bool {
	if withoutHyphen {
		return name == e.Name || e.contains(e.shortnames, name) || e.contains(e.noHyphens, name)
	} else {
		return name == e.Name || e.contains(e.shortnames, name)
	}
}

func (e Entry) Distance(name string, withoutHyphen bool) (int, string) {
	distance := edlib.LevenshteinDistance(name, e.Name)
	command := e.Name
	// TODO shortnameどうする？一番小さいDistanceでいいか？
	if len(e.shortnames) > 0 {
		for _, tmp := range e.shortnames {
			sd := edlib.LevenshteinDistance(name, tmp)
			if sd < distance {
				distance = sd
				command = tmp
			}
		}
	}
	if withoutHyphen && len(e.noHyphens) > 0 {
		for _, tmp := range e.noHyphens {
			sd := edlib.LevenshteinDistance(name, tmp)
			if sd < distance {
				distance = sd
				command = tmp
			}
		}
	}
	return distance, command
}

// slices.Contains support >= Go 1.21
func (e Entry) contains(names []string, target string) bool {
	for _, name := range names {
		if target == name {
			return true
		}
	}
	return false
}

type Commands struct {
	entries []Entry
}

func (c Commands) Find(name string, withoutHyphen bool) (Definition, error) {
	for _, entry := range c.entries {
		if entry.IsTarget(name, withoutHyphen) {
			return entry.definition, nil
		}
	}
	return Definition{}, fmt.Errorf("not found %s command", name)
}

func (c Commands) DidYouMean(name string, withoutHyphen bool) ([]Definition, []string) {
	var candidates []Definition
	var cmds []string
	for _, entry := range c.entries {
		d, cmd := entry.Distance(name, withoutHyphen)
		// TODO 3にした場合は、candidatesの距離の小さい順で返したほうが便利な気がする
		if d < 3 {
			candidates = append(candidates, entry.definition)
			cmds = append(cmds, cmd)
		}
	}
	return candidates, cmds
}

func (c Commands) Help() string {
	var builder strings.Builder
	for _, command := range c.entries {
		builder.WriteString(command.Help)
	}
	return builder.String()
}

type Config struct {
	Owntone   owntone.Config   `json:"Owntone"`
	Switchbot switchbot.Config `json:"Switchbot"`
	Yamaha    yamaha.Config    `json:"Yamaha"`
	Commands  Commands
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
		},
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
