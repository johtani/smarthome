package subcommand

import (
	"github.com/johtani/smarthome/subcommand/action"
	"github.com/johtani/smarthome/subcommand/action/owntone"
	"github.com/johtani/smarthome/subcommand/action/yamaha"
)

// ChangePlaylistCmd is the command name for changing the music playlist.
const ChangePlaylistCmd = "change playlist"

// NewChangePlaylistCmdDefinition creates the definition for the change playlist command.
func NewChangePlaylistCmdDefinition() Definition {
	return Definition{
		Name:        ChangePlaylistCmd,
		Description: "Change Playlist",
		Factory:     NewChangePlaylistSubcommand,
	}
}

// NewChangePlaylistSubcommand creates a new Subcommand for the change playlist command.
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
