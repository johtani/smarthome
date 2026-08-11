package subcommand

import (
	"fmt"
	"strings"
)

const (
	argsValidationReasonValid             = "valid"
	argsValidationReasonUnexpected        = "unexpected_argument"
	argsValidationReasonMissingRequired   = "missing_required_argument"
	argsValidationReasonInvalidEnum       = "invalid_enum_value"
	argsValidationReasonInvalidPrefix     = "invalid_prefix"
	argsValidationReasonDuplicateArgument = "duplicate_argument"
)

// validateResolvedArgs validates natural-language resolver output against the
// command definition. It intentionally requires exact enum values: fuzzy
// matching is useful for direct user input, but must not silently repair model
// output before execution.
func validateResolvedArgs(def Definition, resolvedArgs string) (string, bool) {
	tokens := strings.Fields(strings.TrimSpace(resolvedArgs))
	if len(def.Args) == 0 {
		if len(tokens) == 0 {
			return argsValidationReasonValid, true
		}
		return argsValidationReasonUnexpected, false
	}

	matched := make(map[string]bool, len(def.Args))
	remaining := make([]string, 0, len(tokens))

	for _, token := range tokens {
		matchedPrefix := false
		for _, arg := range def.Args {
			if arg.Prefix == "" || !strings.HasPrefix(token, arg.Prefix) {
				continue
			}
			matchedPrefix = true
			if matched[arg.Name] {
				return argsValidationReasonDuplicateArgument, false
			}
			value := strings.TrimPrefix(token, arg.Prefix)
			if value == "" {
				return argsValidationReasonInvalidPrefix, false
			}
			if !isExactEnumValue(arg, value) {
				return argsValidationReasonInvalidEnum, false
			}
			matched[arg.Name] = true
			break
		}
		if matchedPrefix {
			continue
		}

		for _, arg := range def.Args {
			if arg.Prefix == "" && strings.HasPrefix(token, arg.Name+":") {
				return argsValidationReasonInvalidPrefix, false
			}
		}
		remaining = append(remaining, token)
	}

	freeFormArgs := make([]Arg, 0, 1)
	for _, arg := range def.Args {
		if arg.Prefix != "" {
			if arg.Required && !matched[arg.Name] {
				return argsValidationReasonMissingRequired, false
			}
			continue
		}
		if len(arg.Enum) == 0 {
			freeFormArgs = append(freeFormArgs, arg)
			continue
		}

		matchIndex := -1
		for i, token := range remaining {
			if isExactEnumValue(arg, token) {
				matchIndex = i
				break
			}
		}
		if matchIndex >= 0 {
			matched[arg.Name] = true
			remaining = append(remaining[:matchIndex], remaining[matchIndex+1:]...)
		} else if arg.Required {
			return argsValidationReasonMissingRequired, false
		}
	}

	if len(freeFormArgs) > 1 {
		return fmt.Sprintf("%s:%s", argsValidationReasonUnexpected, "ambiguous_schema"), false
	}
	if len(freeFormArgs) == 1 {
		arg := freeFormArgs[0]
		if len(remaining) > 0 {
			matched[arg.Name] = true
			remaining = nil
		}
		if arg.Required && !matched[arg.Name] {
			return argsValidationReasonMissingRequired, false
		}
	}

	if len(remaining) > 0 {
		return argsValidationReasonInvalidEnum, false
	}
	return argsValidationReasonValid, true
}

func isExactEnumValue(arg Arg, value string) bool {
	if len(arg.Enum) == 0 {
		return true
	}
	for _, allowed := range arg.Enum {
		if value == allowed {
			return true
		}
	}
	return false
}
