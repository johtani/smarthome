package subcommand

import (
	"fmt"
	"github.com/hbollon/go-edlib"
	"github.com/johtani/smarthome/subcommand/action"
	"strings"
	"unicode"
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

func (d Definition) Distance(input string) (int, string) {
	// 対象にならないサイズを初期値として設定
	distance := dymCandidateDistance + 1
	command := ""
	// TODO shortnameどうする？一番小さいDistanceでいいか？
	for _, cmd := range append([]string{d.Name}, d.shortnames...) {
		sd := edlib.LevenshteinDistance(input, cmd)
		if sd < distance {
			distance = sd
			command = cmd
		}
	}
	return distance, command
}

func (d Definition) Match(message string) (bool, string) {
	var args = ""
	inputs := strings.Fields(message)
	for _, cmd := range append([]string{d.Name}, d.shortnames...) {
		cmds := strings.Fields(cmd)
		if d.isPrefix(inputs, cmds) {
			args = message
			for _, cmd := range cmds {
				args = strings.Replace(args, cmd, "", 1)
				args = strings.TrimLeftFunc(args, unicode.IsSpace)
			}
			return true, args
		}
	}
	return false, args
}

func (d Definition) isPrefix(inputs []string, cmd []string) bool {
	if len(inputs) == 0 || len(cmd) == 0 || len(inputs) < len(cmd) {
		return false
	}
	for i := 0; i < len(cmd); i++ {
		if cmd[i] != inputs[i] {
			return false
		}
	}
	return true
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
			NewSearchMusicCmdDefinition(),
			NewSearchAndPlayMusicCmdDefinition(),
			NewSwitchBotDeviceListDefinition(),
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
		find, args = d.Match(text)
		if find {
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

const dymCandidateDistance = 3

func (c Commands) didYouMean(name string) ([]Definition, []string) {
	var candidates []Definition
	var cmds []string
	for _, def := range c.definitions {
		d, cmd := def.Distance(name)
		// TODO 3にした場合は、candidatesの距離の小さい順で返したほうが便利な気がする
		if d < dymCandidateDistance {
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
