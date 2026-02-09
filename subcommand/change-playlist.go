package subcommand

import (
	"github.com/johtani/smarthome/subcommand/action"
	"github.com/johtani/smarthome/subcommand/action/owntone"
	"github.com/johtani/smarthome/subcommand/action/yamaha"
)

const ChangePlaylistCmd = "change playlist"

func NewChangePlaylistCmdDefinition() Definition {
	return Definition{
		Name:        ChangePlaylistCmd,
		Description: "Change Playlist",
		Factory:     NewChangePlaylistSubcommand,
	}
}

func NewChangePlaylistSubcommand(definition Definition, config Config) Subcommand {
	owntoneClient := owntone.NewClient(config.Owntone)
	yamahaClient := yamaha.NewClient(config.Yamaha)
	return Subcommand{
		Definition: definition,
		actions: []action.Action{
			owntone.NewClearQueueAction(owntoneClient),
			yamaha.NewPowerOnAction(yamahaClient),
			yamaha.NewSetInputAction(yamahaClient, "airplay"),
			owntone.NewChangePlaylistAction(owntoneClient),
		},
		ignoreError: true,
	}
}
