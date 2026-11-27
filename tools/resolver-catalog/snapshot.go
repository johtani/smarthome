package main

import (
	"fmt"
	"strings"

	"github.com/johtani/smarthome/subcommand"
)

// Snapshots store Definition.ArgHints rather than live configuration. Read the
// schema that the model saw and leave value validation to ValidateResolvedArgs.
func snapshotDefinitions(catalog []entry) (map[string]subcommand.Definition, error) {
	defs := make(map[string]subcommand.Definition, len(catalog))
	for _, e := range catalog {
		name := strings.TrimSpace(e.Name)
		if _, found := defs[name]; found || name == "" {
			return nil, fmt.Errorf("empty or duplicate snapshot command %q", name)
		}
		d := subcommand.Definition{Name: name}
		hints := strings.TrimSpace(e.Args)
		for hints != "" {
			spec, rest, ok := strings.Cut(hints, ")")
			argName, options, hasOptions := strings.Cut(spec, "(")
			if !ok || !hasOptions || argName == "" {
				return nil, fmt.Errorf("invalid argument hint for %q: %q", name, hints)
			}
			arg := subcommand.Arg{Name: argName}
			for option := range strings.SplitSeq(options, ",") {
				switch {
				case option == "required":
					arg.Required = true
				case option == "optional", option == "no-prefix":
				case strings.HasPrefix(option, "prefix="):
					arg.Prefix = strings.TrimPrefix(option, "prefix=")
				case strings.HasPrefix(option, "enum="):
					arg.Enum = strings.Split(strings.TrimPrefix(option, "enum="), "|")
				default:
					return nil, fmt.Errorf("unknown argument option for %q: %q", name, option)
				}
			}
			d.Args = append(d.Args, arg)
			if rest == "" {
				break
			}
			next, ok := strings.CutPrefix(strings.TrimSpace(rest), ",")
			hints = strings.TrimSpace(next)
			if !ok || hints == "" {
				return nil, fmt.Errorf("invalid argument separator for %q", name)
			}
		}
		defs[name] = d
	}
	return defs, nil
}
