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
	WithArgs    bool
	Factory     func(Definition, Config) Subcommand
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

func (c Commands) Find(name string, withoutHyphen bool) (Definition, string, string, error) {
	var d Definition
	var args string
	dymMsg := ""
	find := false
	for _, entry := range c.entries {
		if entry.definition.WithArgs {
			params := strings.SplitN(name, " ", 2)
			if len(params) < 2 {
				return Definition{}, "", "", fmt.Errorf("%s is not supported without arguments", name)
			}
			if entry.IsTarget(params[0], withoutHyphen) {
				d = entry.definition
				args = params[1]
				find = true
				break
			}
		} else {
			if entry.IsTarget(name, withoutHyphen) {
				d = entry.definition
				args = ""
				find = true
				break
			}
		}
	}

	if find == false {
		candidates, cmds := c.didYouMean(name, true)
		if len(candidates) == 0 {
			return Definition{}, "", "", fmt.Errorf("Sorry, I cannot understand what you want from what you said '%v'...\n", name)
		} else {
			d = candidates[0]
			dymMsg = fmt.Sprintf("Did you mean \"%v\"?", cmds[0])
		}
	}

	return d, args, dymMsg, nil
}

func (c Commands) didYouMean(name string, withoutHyphen bool) ([]Definition, []string) {
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
