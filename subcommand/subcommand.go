package subcommand

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"unicode"

	"github.com/hbollon/go-edlib"
	"github.com/johtani/smarthome/subcommand/action"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type Subcommand struct {
	Definition
	actions     []action.Action
	ignoreError bool
}

func (s Subcommand) Exec(ctx context.Context, args string) (string, error) {
	ctx, span := otel.Tracer("subcommand").Start(ctx, "Subcommand.Exec")
	defer span.End()
	span.SetAttributes(
		attribute.String("subcommand.name", s.Name),
		attribute.String("subcommand.args", args),
	)
	var msgs []string
	for i := range s.actions {
		msg, err := s.actions[i].Run(ctx, args)
		if s.ignoreError && err != nil {
			msg = fmt.Sprintf("skip error\t %v\n", err)
			msgs = append(msgs, msg)
		} else if err != nil {
			return "", err
		} else {
			msgs = append(msgs, msg)
		}
	}
	return strings.Join(msgs, "\n"), nil
}

type Definition struct {
	Name        string
	Description string
	shortnames  []string
	Factory     func(Definition, Config) Subcommand
	Args        []Arg
}

type Arg struct {
	Name        string
	Description string
	Required    bool
	Enum        []string
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
	Definitions []Definition
}

func NewCommands() Commands {
	return Commands{
		Definitions: []Definition{
			NewStartMeetingDefinition(),
			NewFinishMeetingDefinition(),
			NewStartMusicCmdDefinition(),
			NewStopMusicDefinition(),
			NewChangePlaylistCmdDefinition(),
			NewDisplayPalylistCmdDefinition(),
			NewDisplayOutputsCmdDefinition(),
			NewUpdateLibraryCmdDefinition(),
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
	for _, d := range c.Definitions {
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

type dymCandidate struct {
	def      Definition
	cmd      string
	distance int
}

func (c Commands) didYouMean(name string) ([]Definition, []string) {
	var candidates []dymCandidate
	for _, def := range c.Definitions {
		d, cmd := def.Distance(name)
		if d < dymCandidateDistance {
			candidates = append(candidates, dymCandidate{def, cmd, d})
		}
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].distance < candidates[j].distance
	})

	var resultDefs []Definition
	var resultCmds []string
	for _, candidate := range candidates {
		resultDefs = append(resultDefs, candidate.def)
		resultCmds = append(resultCmds, candidate.cmd)
	}
	return resultDefs, resultCmds
}

func (c Commands) Help() string {
	var builder strings.Builder
	for _, d := range c.Definitions {
		builder.WriteString(d.Help())
	}
	return builder.String()
}
