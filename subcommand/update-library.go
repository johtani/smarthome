package subcommand

import (
	"github.com/johtani/smarthome/subcommand/action"
	"github.com/johtani/smarthome/subcommand/action/owntone"
)

// UpdateLibraryCmd is the command name for updating the music library.
const UpdateLibraryCmd = "update library"

// NewUpdateLibraryCmdDefinition creates the definition for the update library command.
func NewUpdateLibraryCmdDefinition() Definition {
	return Definition{
		Name:        UpdateLibraryCmd,
		Description: "Update library",
		Factory:     NewUpdateLibrarySubcommand,
	}
}

// NewUpdateLibrarySubcommand creates a new Subcommand for the update library command.
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
