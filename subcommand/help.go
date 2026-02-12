package subcommand

import (
	"github.com/johtani/smarthome/subcommand/action"
)

// HelpCmd is the command name for displaying help.
const HelpCmd = "help"

// NewHelpDefinition creates the definition for the help command.
func NewHelpDefinition() Definition {
	return Definition{
		Name:        HelpCmd,
		Description: "Display commands list",
		Factory:     NewHelpSubcommand,
	}
}

// NewHelpSubcommand creates a new Subcommand for the help command.
func NewHelpSubcommand(definition Definition, config Config) Subcommand {
	return Subcommand{
		Definition: definition,
		actions: []action.Action{
			action.NewHelpAction(config.Commands.Help()),
		},
	}
}
