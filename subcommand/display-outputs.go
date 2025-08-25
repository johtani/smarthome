package subcommand

import (
	"github.com/johtani/smarthome/subcommand/action"
	"github.com/johtani/smarthome/subcommand/action/owntone"
)

const DisplayOutputsCmd = "display outputs"
const DispOutputsCmd = "disp outputs"

func NewDisplayOutputsCmdDefinition() Definition {
	return Definition{
		Name:        DisplayOutputsCmd,
		shortnames:  []string{DispOutputsCmd},
		Description: "Display outputs (Owntone)",
		Factory:     NewDisplayOutputsSubcommand,
	}
}

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
