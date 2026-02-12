/*
Package subcommand provides the infrastructure for defining and executing smart home commands.
Each command is composed of one or more actions.
*/
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

// Subcommand represents a executable command consisting of one or more actions.
type Subcommand struct {
	Definition
	actions     []action.Action
	ignoreError bool
}

// Exec executes the subcommand by running all its actions in sequence.
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
		switch {
		case s.ignoreError && err != nil:
			msg = fmt.Sprintf("skip error\t %v\n", err)
			msgs = append(msgs, msg)
		case err != nil:
			return "", err
		default:
			msgs = append(msgs, msg)
		}
	}
	return strings.Join(msgs, "\n"), nil
}

// Definition defines the metadata and factory for a subcommand.
type Definition struct {
	Name        string
	Description string
	shortnames  []string
	Factory     func(Definition, Config) Subcommand
	Args        []Arg
}

// Arg represents an argument for a subcommand.
type Arg struct {
	Name        string
	Description string
	Required    bool
	Enum        []string
	Prefix      string
}

// Init initializes a subcommand from its definition and configuration.
func (d Definition) Init(config Config) Subcommand {
	return d.Factory(d, config)
}

// Help returns a help string for the subcommand.
func (d Definition) Help() string {
	var help string
	if len(d.shortnames) > 0 {
		help = fmt.Sprintf("  %s [%s]: %s\n", d.Name, strings.Join(d.shortnames, "/"), d.Description)
	} else {
		help = fmt.Sprintf("  %s : %s\n", d.Name, d.Description)
	}
	return help
}

// Distance calculates the Levenshtein distance between the input and the command names.
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

// Match checks if the given message matches the subcommand name or its shortnames.
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

// Commands represents a collection of subcommand definitions.
type Commands struct {
	Definitions []Definition
}

// NewCommands creates a new collection of all available subcommand definitions.
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

// Find searches for a subcommand definition that matches the given text.
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

	if !find {
		candidates, cmds := c.didYouMean(text)
		if len(candidates) == 0 {
			return Definition{}, "", "", fmt.Errorf("sorry, i cannot understand what you want from what you said '%v'", text)
		}
		def = candidates[0]
		dymMsg = fmt.Sprintf("Did you mean \"%v\"?", cmds[0])
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

// Help returns a help string for all subcommands in the collection.
func (c Commands) Help() string {
	var builder strings.Builder
	for _, d := range c.Definitions {
		builder.WriteString(d.Help())
	}
	return builder.String()
}
