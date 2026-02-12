package subcommand

import (
	"github.com/johtani/smarthome/subcommand/action"
	"github.com/johtani/smarthome/subcommand/action/owntone"
)

// DisplayPlaylistsCmd is the command name for displaying playlists.
const DisplayPlaylistsCmd = "display playlist"

// DispPlaylistsCmd is a short name for the display playlists command.
const DispPlaylistsCmd = "disp playlists"

// NewDisplayPalylistCmdDefinition creates the definition for the display playlists command.
func NewDisplayPalylistCmdDefinition() Definition {
	return Definition{
		Name:        DisplayPlaylistsCmd,
		shortnames:  []string{DispPlaylistsCmd},
		Description: "Display playlists",
		Factory:     NewDisplayPalylistsSubcommand,
	}
}

// NewDisplayPalylistsSubcommand creates a new Subcommand for the display playlists command.
func NewDisplayPalylistsSubcommand(definition Definition, config Config) Subcommand {
	owntoneClient := owntone.NewClient(config.Owntone)
	return Subcommand{
		Definition: definition,
		actions: []action.Action{
			owntone.NewDisplayPlaylistsAction(owntoneClient),
		},
		ignoreError: true,
	}
}
