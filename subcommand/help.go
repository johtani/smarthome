package subcommand

import (
	"github.com/johtani/smarthome/subcommand/action"
)

const HelpCmd = "help"

func NewHelpDefinition() Definition {
	return Definition{
		Name:        HelpCmd,
		Description: "Display commands list",
		Factory:     NewHelpSubcommand,
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
