package subcommand

import (
	"fmt"
	"github.com/johtani/smarthome/subcommand/action"
	"strings"
)

const HelpCmd = "help"

func NewHelpDefinition() Definition {
	return Definition{
		HelpCmd,
		"Display commands list",
		NewHelpSubcommand,
	}
}

func NewHelpSubcommand(definition Definition, config Config) Subcommand {
	var builder strings.Builder
	for _, command := range Map() {
		builder.WriteString(fmt.Sprintf("  %s: %s\n", command.Name, command.Description))
	}
	return Subcommand{
		definition,
		[]action.Action{
			action.NewHelpAction(builder.String()),
		},
		false,
	}
}
