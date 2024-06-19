package subcommand

import (
	"github.com/johtani/smarthome/subcommand/action"
	"github.com/johtani/smarthome/subcommand/action/owntone"
)

const UpdateLibraryCmd = "update library"

func NewUpdateLibraryCmdDefinition() Definition {
	return Definition{
		Name:        UpdateLibraryCmd,
		Description: "Update library",
		Factory:     NewUpdateLibrarySubcommand,
	}
}

func NewUpdateLibrarySubcommand(definition Definition, config Config) Subcommand {
	owntoneClient := owntone.NewClient(config.Owntone)
	return Subcommand{
		Definition: definition,
		actions: []action.Action{
			owntone.NewUpdateLibraryAction(owntoneClient),
		},
		ignoreError: true,
	}
}
