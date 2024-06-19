package subcommand

import (
	"github.com/johtani/smarthome/subcommand/action"
	"github.com/johtani/smarthome/subcommand/action/owntone"
)

const DisplayPlaylistsCmd = "display playlist"
const DispPlaylistsCmd = "disp playlists"

func NewDisplayPalylistCmdDefinition() Definition {
	return Definition{
		Name:        DisplayPlaylistsCmd,
		shortnames:  []string{DispPlaylistsCmd},
		Description: "Display playlists",
		Factory:     NewDisplayPalylistsSubcommand,
	}
}

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
