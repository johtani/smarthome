package subcommand

import (
	"github.com/johtani/smarthome/subcommand/action"
	"github.com/johtani/smarthome/subcommand/action/owntone"
)

const ChangePlaylistCmd = "change-playlist"

func NewChangePlaylistCmdDefinition() Definition {
	return Definition{
		ChangePlaylistCmd,
		"Change Playlist",
		NewChangePlaylistSubcommand,
	}
}

func NewChangePlaylistSubcommand(definition Definition, config Config) Subcommand {
	owntoneClient := owntone.NewClient(config.Owntone)
	return Subcommand{
		definition,
		[]action.Action{
			owntone.NewChangePlaylistAction(owntoneClient),
		},
		true,
	}
}
