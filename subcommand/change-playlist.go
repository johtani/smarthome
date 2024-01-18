package subcommand

import (
	"github.com/johtani/smarthome/subcommand/action"
	"github.com/johtani/smarthome/subcommand/action/owntone"
)

const ChangePlaylistCmd = "change-playlist"

func NewChangePlaylistCmdDefinition() Definition {
	return Definition{
		Name:        ChangePlaylistCmd,
		Description: "Change Playlist",
		Factory:     NewChangePlaylistSubcommand,
	}
}

func NewChangePlaylistSubcommand(definition Definition, config Config) Subcommand {
	owntoneClient := owntone.NewClient(config.Owntone)
	return Subcommand{
		Definition: definition,
		actions: []action.Action{
			owntone.NewChangePlaylistAction(owntoneClient),
		},
		ignoreError: true,
	}
}
