package subcommand

import (
	"github.com/johtani/smarthome/subcommand/action"
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
	return Subcommand{
		definition,
		[]action.Action{
			action.NewHelpAction(config.Commands.Help()),
		},
		false,
	}
}
