// Command resolver-catalog snapshots runtime definitions and checks labels without device I/O.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/johtani/smarthome/subcommand"
)

type entry struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Args        string `json:"args"`
}

type label struct {
	Command string `json:"command"`
	Args    string `json:"args"`
}

func run() error {
	dir := flag.String("config-dir", "config", "directory containing macros.json")
	snapshot := flag.Bool("snapshot", false, "validate a batch against a supplied catalog snapshot on stdin")
	flag.Parse()
	if *snapshot {
		var batch struct {
			Catalog []entry `json:"catalog"`
			Labels  []label `json:"labels"`
		}
		if err := json.NewDecoder(os.Stdin).Decode(&batch); err != nil {
			return err
		}
		defs, err := snapshotDefinitions(batch.Catalog)
		if err != nil {
			return err
		}
		return json.NewEncoder(os.Stdout).Encode(validateLabels(defs, batch.Labels))
	}
	commands, err := subcommand.LoadCommandDefinitions(*dir)
	if err != nil {
		return err
	}
	var labels []label
	if err := json.NewDecoder(os.Stdin).Decode(&labels); err != nil {
		return err
	}
	defs := make(map[string]subcommand.Definition)
	var catalog []entry
	for _, d := range commands.Definitions {
		defs[d.Name] = d
		catalog = append(catalog, entry{d.Name, d.Description, d.ArgHints(true)})
	}
	return json.NewEncoder(os.Stdout).Encode(map[string]any{"catalog": catalog, "errors": validateLabels(defs, labels)})
}

func validateLabels(defs map[string]subcommand.Definition, labels []label) []string {
	reasons := make([]string, len(labels))
	for i, l := range labels {
		switch {
		case strings.TrimSpace(l.Command) == "":
			if l.Command != "" || l.Args != "" {
				reasons[i] = "invalid_unresolved"
			}
		default:
			d, ok := defs[l.Command]
			if !ok {
				reasons[i] = "unknown_command"
				continue
			}
			if reason, valid := d.ValidateResolvedArgs(l.Args); !valid {
				reasons[i] = reason
			}
		}
	}
	return reasons
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
