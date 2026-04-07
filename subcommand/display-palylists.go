package subcommand

import (
	"github.com/johtani/smarthome/subcommand/action"
	"github.com/johtani/smarthome/subcommand/action/owntone"
)

// DisplayPlaylistsCmd is the command name for displaying playlists.
const DisplayPlaylistsCmd = "display playlist"

// DispPlaylistsCmd is a short name for the display playlists command.
const DispPlaylistsCmd = "disp playlists"

// NewDisplayPlaylistCmdDefinition creates the definition for the display playlists command.
func NewDisplayPlaylistCmdDefinition() Definition {
	return Definition{
		Name:        DisplayPlaylistsCmd,
		shortnames:  []string{DispPlaylistsCmd},
		Description: "Display playlists",
		Factory:     NewDisplayPlaylistsSubcommand,
	}
}

// NewDisplayPlaylistsSubcommand creates a new Subcommand for the display playlists command.
func NewDisplayPlaylistsSubcommand(definition Definition, config Config) Subcommand {
	owntoneClient := owntone.NewClient(config.Owntone)
	return Subcommand{
		Definition: definition,
		actions: []action.Action{
			owntone.NewDisplayPlaylistsAction(owntoneClient),
		},
		ignoreError: true,
	}
}
