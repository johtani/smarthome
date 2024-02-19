package subcommand

import (
	"fmt"
	"github.com/hbollon/go-edlib"
	"github.com/johtani/smarthome/subcommand/action"
	"strings"
)

type Subcommand struct {
	Definition
	actions     []action.Action
	ignoreError bool
}

func (s Subcommand) Exec(args string) (string, error) {
	var msgs []string
	for i := range s.actions {
		msg, err := s.actions[i].Run(args)
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

type Definition struct {
	Name        string
	Description string
	shortnames  []string
	WithArgs    bool
	Factory     func(Definition, Config) Subcommand
}

func (d Definition) Init(config Config) Subcommand {
	return d.Factory(d, config)
}

func (d Definition) Help() string {
	var help string
	if len(d.shortnames) > 0 {
		help = fmt.Sprintf("  %s [%s]: %s\n", d.Name, strings.Join(d.shortnames, "/"), d.Description)
	} else {
		help = fmt.Sprintf("  %s : %s\n", d.Name, d.Description)
	}
	return help
}

func (d Definition) noHyphens() []string {
	noHyphens := []string{strings.ReplaceAll(d.Name, "-", " ")}
	for _, shortname := range d.shortnames {
		noHyphens = append(noHyphens, strings.ReplaceAll(shortname, "-", " "))
	}
	return noHyphens
}

func (d Definition) Distance(name string) (int, string) {
	distance := edlib.LevenshteinDistance(name, d.Name)
	command := d.Name
	// TODO shortnameどうする？一番小さいDistanceでいいか？
	if len(d.shortnames) > 0 {
		for _, tmp := range d.shortnames {
			sd := edlib.LevenshteinDistance(name, tmp)
			if sd < distance {
				distance = sd
				command = tmp
			}
		}
	}
	return distance, command
}

func (d Definition) Match(message string) (bool, string, error) {
	var match bool = false
	var args string = ""
	idx := d.lastPositionOfName(message)
	if idx > -1 {
		match = true
		if d.WithArgs && len(message) >= idx+1 {
			args = message[idx+1:]
		}
	}
	return match, args, nil
}

func DefaultMatch(message string) (bool, string) {
	return false, ""
}

func (d Definition) lastPositionOfName(message string) int {
	var idx int = -1
	if strings.HasPrefix(message, d.Name) {
		return len(d.Name)
	}
	for _, shortname := range d.shortnames {
		if strings.HasPrefix(message, shortname) {
			return len(shortname)
		}
	}
	return idx
}

type Commands struct {
	definitions []Definition
}

func NewCommands() Commands {
	return Commands{
		definitions: []Definition{
			NewStartMeetingDefinition(),
			NewFinishMeetingDefinition(),
			NewStartMusicCmdDefinition(),
			NewStopMusicDefinition(),
			NewChangePlaylistCmdDefinition(),
			NewSwitchBotDeviceListDefinition(),
			NewSwitchBotDeviceListDefinition(),
			NewSwitchBotSceneListDefinition(),
			NewSwitchBotSceneListDefinition(),
			NewLightOffDefinition(),
			NewLightOnDefinition(),
			NewHelpDefinition(),
			NewStartSwitchDefinition(),
			NewStartPS5Definition(),
			NewAirConditionerOnDefinition(),
			NewAirConditionerOffDefinition(),
			NewDisplayTemperatureDefinition(),
			NewTokenizeIpaDefinition(),
			NewTokenizeUniDefinition(),
			NewTokenizeNeologdDefinition(),
		},
	}
}

func (c Commands) Find(text string) (Definition, string, string, error) {
	var def Definition
	var args string
	dymMsg := ""
	find := false
	for _, d := range c.definitions {
		var err error
		find, args, err = d.Match(text)
		if err != nil {
			return Definition{}, "", "", err
		} else if find {
			def = d
			break
		}
	}

	if find == false {
		candidates, cmds := c.didYouMean(text)
		if len(candidates) == 0 {
			return Definition{}, "", "", fmt.Errorf("Sorry, I cannot understand what you want from what you said '%v'...\n", text)
		} else {
			def = candidates[0]
			dymMsg = fmt.Sprintf("Did you mean \"%v\"?", cmds[0])
		}
	}

	return def, args, dymMsg, nil
}

func (c Commands) didYouMean(name string) ([]Definition, []string) {
	var candidates []Definition
	var cmds []string
	for _, def := range c.definitions {
		d, cmd := def.Distance(name)
		// TODO 3にした場合は、candidatesの距離の小さい順で返したほうが便利な気がする
		if d < 3 {
			candidates = append(candidates, def)
			cmds = append(cmds, cmd)
		}
	}
	return candidates, cmds
}

func (c Commands) Help() string {
	var builder strings.Builder
	for _, d := range c.definitions {
		builder.WriteString(d.Help())
	}
	return builder.String()
}
