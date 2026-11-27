package subcommand

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"sort"
	"strings"

	"github.com/johtani/smarthome/internal/resolver"
	"github.com/johtani/smarthome/subcommand/action/llm"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

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
	lastResolverResult := llm.ResolvedCommand{PromptVersion: config.Resolver.PromptVersion}
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
				PromptVersion:     config.Resolver.PromptVersion,
				ResolvedCommand:   def.Name,
				ResolvedArgs:      args,
				DidYouMeanCommand: cmds[0],
			})
			return def, args, dymMsg, nil
		}

	resolverLoop:
		for _, nlr := range naturalLanguageResolvers(config) {
			span.SetAttributes(attribute.String("resolver.path", nlr.Path()))
			resolved, err := nlr.Resolve(ctx, text, c.Help(), config.Resolver.PromptVersion)
			if err != nil {
				slog.WarnContext(ctx, "natural language resolver failed", "resolver", nlr.Path(), "error", err)
				span.RecordError(err)
				continue
			}
			lastResolverResult = resolved
			if resolved.Command == "" {
				continue
			}

			initialResolution := resolved
			resolved, d, validationReason, argsValid := c.checkResolverOutput(text, resolved)
			correction := describeResolutionCorrection(initialResolution, resolved)

			if d.Name != "" {
				span.AddEvent("resolver.args.validation", trace.WithAttributes(
					attribute.String("resolver.path", nlr.Path()),
					attribute.String("resolver.resolved_command", resolved.Command),
					attribute.Bool("resolver.args.valid", argsValid),
					attribute.String("resolver.args.validation_reason", validationReason),
				))
				span.SetAttributes(
					attribute.Bool("resolver.args.valid", argsValid),
					attribute.String("resolver.args.validation_reason", validationReason),
				)
				if !argsValid {
					slog.WarnContext(
						ctx,
						"natural language resolver returned invalid arguments",
						"resolver", nlr.Path(),
						"command", resolved.Command,
						"reason", validationReason,
					)
					continue resolverLoop
				}
				span.SetAttributes(
					attribute.String("resolver.initial_command", correction.initialCommand),
					attribute.String("resolver.initial_args_kind", correction.initialArgsKind),
					attribute.Bool("resolver.command_corrected", correction.commandCorrected),
					attribute.String("resolver.correction_reason", correction.reason),
					attribute.String("resolver.resolved_command", resolved.Command),
					attribute.String("resolver.resolved_args", resolved.Args),
				)
				resolver.RecordDecision(ctx, resolver.DecisionRecord{
					InputTextHash:    inputHash,
					ResolverPath:     nlr.Path(),
					ResolverMode:     config.Resolver.Mode,
					LLMModel:         resolved.Model,
					PromptVersion:    resolved.PromptVersion,
					ArtifactVersion:  resolved.ArtifactVersion,
					DatasetVersion:   resolved.DatasetVersion,
					InitialCommand:   correction.initialCommand,
					InitialArgsKind:  correction.initialArgsKind,
					CommandCorrected: correction.commandCorrected,
					CorrectionReason: correction.reason,
					ResolvedCommand:  resolved.Command,
					ResolvedArgs:     resolved.Args,
				})
				return d, resolved.Args, fmt.Sprintf("(%s) %s", nlr.Prefix(), resolved.Thought), nil
			}
			slog.WarnContext(ctx, "resolver returned an unknown command", "resolver", nlr.Path(), "command", resolved.Command)
			span.SetAttributes(attribute.Bool("resolver.nl_unknown_command", true))
		}

		err := fmt.Errorf("sorry, i cannot understand what you want from what you said '%v'", text)
		span.RecordError(err)
		span.SetAttributes(attribute.String("resolver.path", "unresolved"))
		resolver.RecordDecision(ctx, resolver.DecisionRecord{
			InputTextHash:   inputHash,
			ResolverPath:    "unresolved",
			ResolverMode:    config.Resolver.Mode,
			LLMModel:        lastResolverResult.Model,
			PromptVersion:   lastResolverResult.PromptVersion,
			ArtifactVersion: lastResolverResult.ArtifactVersion,
			DatasetVersion:  lastResolverResult.DatasetVersion,
		})
		return Definition{}, "", "", err
	}

	resolver.RecordDecision(ctx, resolver.DecisionRecord{
		InputTextHash:   inputHash,
		ResolverPath:    "exact_match",
		ResolverMode:    config.Resolver.Mode,
		PromptVersion:   config.Resolver.PromptVersion,
		ResolvedCommand: def.Name,
		ResolvedArgs:    args,
	})
	return def, args, dymMsg, nil
}

func stripNaturalLanguageSearchTypeArgs(args string) string {
	allowedTypes := map[string]struct{}{
		"artist": {},
		"album":  {},
		"track":  {},
		"genre":  {},
	}
	parts := strings.Fields(strings.ReplaceAll(args, ",", " "))
	kept := make([]string, 0, len(parts))
	for _, part := range parts {
		value, ok := strings.CutPrefix(part, "type:")
		if ok {
			if _, allowed := allowedTypes[value]; allowed {
				continue
			}
		}
		kept = append(kept, part)
	}
	return strings.Join(kept, " ")
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
