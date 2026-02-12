package subcommand

import (
	"github.com/johtani/smarthome/subcommand/action"
	"github.com/johtani/smarthome/subcommand/action/owntone"
)

// DisplayOutputsCmd is the command name for displaying audio outputs.
const DisplayOutputsCmd = "display outputs"

// DispOutputsCmd is a short name for the display outputs command.
const DispOutputsCmd = "disp outputs"

// NewDisplayOutputsCmdDefinition creates the definition for the display outputs command.
func NewDisplayOutputsCmdDefinition() Definition {
	return Definition{
		Name:        DisplayOutputsCmd,
		shortnames:  []string{DispOutputsCmd},
		Description: "Display outputs (Owntone)",
		Factory:     NewDisplayOutputsSubcommand,
	}
}

// NewDisplayOutputsSubcommand creates a new Subcommand for the display outputs command.
func NewDisplayOutputsSubcommand(definition Definition, config Config) Subcommand {
	owntoneClient := owntone.NewClient(config.Owntone)
	return Subcommand{
		Definition: definition,
		actions: []action.Action{
			owntone.NewDisplayOutputsAction(owntoneClient),
		},
		ignoreError: true,
	}
}
