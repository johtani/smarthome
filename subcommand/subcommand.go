/*
Package subcommand provides the infrastructure for defining and executing smart home commands.
Each command is composed of one or more actions.
*/
package subcommand

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"unicode"

	"github.com/hbollon/go-edlib"
	"github.com/johtani/smarthome/internal/resolver"
	"github.com/johtani/smarthome/subcommand/action"
	"github.com/johtani/smarthome/subcommand/action/llm"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
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

// Match checks if the given input matches the argument's prefix and enum values.
func (a Arg) Match(input string) (string, bool) {
	target := input
	if a.Prefix != "" {
		if strings.HasPrefix(input, a.Prefix) {
			target = input[len(a.Prefix):]
		} else {
			return "", false
		}
	}

	if len(a.Enum) == 0 {
		return target, true
	}

	for _, e := range a.Enum {
		if e == target {
			return e, true
		}
	}

	// Try fuzzy match for Enum
	res, err := edlib.FuzzySearch(target, a.Enum, edlib.Levenshtein)
	if err == nil && res != "" {
		// 距離が離れすぎている場合はマッチとみなさない
		distance := edlib.LevenshteinDistance(target, res)
		if distance <= 2 {
			return res, true
		}
	}

	return "", false
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
	if len(d.Args) > 0 {
		var items []string
		for _, arg := range d.Args {
			var details []string
			if arg.Required {
				details = append(details, "required")
			} else {
				details = append(details, "optional")
			}
			if arg.Prefix != "" {
				details = append(details, "prefix="+arg.Prefix)
			}
			if len(arg.Enum) > 0 {
				details = append(details, "enum="+strings.Join(arg.Enum, "|"))
			}
			items = append(items, fmt.Sprintf("%s(%s)", arg.Name, strings.Join(details, ",")))
		}
		help += fmt.Sprintf("    args: %s\n", strings.Join(items, ", "))
	}
	return help
}

// Distance calculates the Levenshtein distance between the input and the command names.
func (d Definition) Distance(input string) (int, string) {
	// 対象にならないサイズを初期値として設定
	distance := dymCandidateDistance + 1
	command := ""
	// TODO shortnameどうする？一番小さいDistanceでいいか？
	inputs := strings.Fields(input)
	for _, cmd := range append([]string{d.Name}, d.shortnames...) {
		cmds := strings.Fields(cmd)
		if len(inputs) < len(cmds) {
			continue
		}
		target := strings.Join(inputs[:len(cmds)], " ")
		sd := edlib.LevenshteinDistance(target, cmd)
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
	for i := range cmd {
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
// Optionally pass MacroConfig values to register user-defined macros.
func NewCommands(macros ...MacroConfig) Commands {
	defs := []Definition{
		NewStartMusicCmdDefinition(),
		NewPlayRandomPlaylistCmdDefinition(),
		NewPlayRandomArtistCmdDefinition(),
		NewPlayRandomGenreCmdDefinition(),
		NewStopMusicDefinition(),
		NewChangePlaylistCmdDefinition(),
		NewDisplayPlaylistCmdDefinition(),
		NewDisplayOutputsCmdDefinition(),
		NewUpdateLibraryCmdDefinition(),
		NewSearchMusicCmdDefinition(),
		NewSearchAndPlayMusicCmdDefinition(),
		NewSwitchBotDeviceListDefinition(),
		NewSwitchBotSceneListDefinition(),
		NewHelpDefinition(),
		NewHealthDefinition(),
		NewDisplayTemperatureDefinition(),
		NewTokenizeIpaDefinition(),
		NewTokenizeUniDefinition(),
		NewTokenizeNeologdDefinition(),
	}
	existingNames := make(map[string]struct{})
	for _, d := range defs {
		existingNames[d.Name] = struct{}{}
	}

	for _, macro := range macros {
		if _, exists := existingNames[macro.Name]; exists {
			slog.Warn("macro skipped: name already registered", "macro_name", macro.Name)
			continue
		}
		defs = append(defs, newMacroDefinition(macro))
		existingNames[macro.Name] = struct{}{}
	}
	return Commands{Definitions: defs}
}

// Find searches for a subcommand definition that matches the given text.
func (c Commands) Find(ctx context.Context, config Config, text string) (Definition, string, string, error) {
	ctx, span := otel.Tracer("resolver").Start(ctx, "Commands.Find", trace.WithAttributes(
		attribute.String("resolver.input_text_hash", hashInputText(text)),
		attribute.String("resolver.mode", config.Resolver.Mode),
	))
	defer span.End()
	if requestID, ok := resolver.RequestIDFromContext(ctx); ok {
		span.SetAttributes(attribute.String("resolver.request_id", requestID))
	}
	if channel, ok := resolver.ChannelFromContext(ctx); ok {
		span.SetAttributes(attribute.String("resolver.channel", channel))
	}

	var def Definition
	var args string
	dymMsg := ""
	inputHash := hashInputText(text)
	find := false
	for _, d := range c.Definitions {
		find, args = d.Match(text)
		if find {
			def = d
			span.SetAttributes(
				attribute.String("resolver.path", "exact_match"),
				attribute.String("resolver.resolved_command", def.Name),
				attribute.String("resolver.resolved_args", args),
			)
			break
		}
	}

	if !find {
		candidates, cmds := c.didYouMean(text)
		if len(candidates) > 0 {
			def = candidates[0]
			dymMsg = fmt.Sprintf("Did you mean \"%v\"?", cmds[0])

			inputs := strings.Fields(text)
			cmdsFields := strings.Fields(cmds[0])
			args = strings.Join(inputs[len(cmdsFields):], " ")
			span.SetAttributes(
				attribute.String("resolver.path", "did_you_mean"),
				attribute.String("resolver.did_you_mean_command", cmds[0]),
				attribute.String("resolver.resolved_command", def.Name),
				attribute.String("resolver.resolved_args", args),
			)
			resolver.RecordDecision(ctx, resolver.DecisionRecord{
				InputTextHash:     inputHash,
				ResolverPath:      "did_you_mean",
				ResolverMode:      config.Resolver.Mode,
				LLMModel:          config.LLM.Model,
				ResolvedCommand:   def.Name,
				ResolvedArgs:      args,
				DidYouMeanCommand: cmds[0],
			})
			return def, args, dymMsg, nil
		}

		// LLMによる解決を試みる
		if config.LLM.Endpoint != "" {
			span.SetAttributes(attribute.String("resolver.path", "llm"))
			client := llm.NewClient(config.LLM)
			resolved, err := client.Resolve(ctx, text, c.Help(), config.Resolver.PromptVersion)
			if err == nil && resolved.Command != "" {
				// Backward-compatible safety fallback:
				// if LLM resolves to start music with free-text args, use search and play.
				if resolved.Command == StartMusicCmd &&
					strings.TrimSpace(resolved.Args) != "" &&
					!strings.HasPrefix(strings.TrimSpace(resolved.Args), "artist") &&
					!strings.HasPrefix(strings.TrimSpace(resolved.Args), "genre") {
					resolved.Command = SearchAndPlayMusicCmd
				}
				for _, d := range c.Definitions {
					if d.Name == resolved.Command {
						span.SetAttributes(
							attribute.String("resolver.resolved_command", resolved.Command),
							attribute.String("resolver.resolved_args", resolved.Args),
						)
						resolver.RecordDecision(ctx, resolver.DecisionRecord{
							InputTextHash:   inputHash,
							ResolverPath:    "llm",
							ResolverMode:    config.Resolver.Mode,
							LLMModel:        config.LLM.Model,
							ResolvedCommand: resolved.Command,
							ResolvedArgs:    resolved.Args,
						})
						return d, resolved.Args, fmt.Sprintf("(LLM) %s", resolved.Thought), nil
					}
				}
				slog.WarnContext(ctx, "LLM resolved to an unknown command", "command", resolved.Command)
				span.SetAttributes(attribute.Bool("resolver.llm_unknown_command", true))
			} else if err != nil {
				slog.ErrorContext(ctx, "LLM resolution failed", "error", err)
				span.RecordError(err)
			}
		}

		err := fmt.Errorf("sorry, i cannot understand what you want from what you said '%v'", text)
		span.RecordError(err)
		span.SetAttributes(attribute.String("resolver.path", "unresolved"))
		resolver.RecordDecision(ctx, resolver.DecisionRecord{
			InputTextHash: inputHash,
			ResolverPath:  "unresolved",
			ResolverMode:  config.Resolver.Mode,
			LLMModel:      config.LLM.Model,
		})
		return Definition{}, "", "", err
	}

	resolver.RecordDecision(ctx, resolver.DecisionRecord{
		InputTextHash:   inputHash,
		ResolverPath:    "exact_match",
		ResolverMode:    config.Resolver.Mode,
		LLMModel:        config.LLM.Model,
		ResolvedCommand: def.Name,
		ResolvedArgs:    args,
	})
	return def, args, dymMsg, nil
}

func hashInputText(text string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(text)))
	return hex.EncodeToString(sum[:12])
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
