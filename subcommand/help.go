package subcommand

import (
	"fmt"
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
	helpMessage := fmt.Sprintf("利用可能なコマンドは次の通りです\n%scommit: %s\n", config.Commands.Help(), currentRevision())
	return Subcommand{
		Definition: definition,
		actions: []action.Action{
			action.NewHelpAction(helpMessage),
		},
	}
}
