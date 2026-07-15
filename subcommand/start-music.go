package subcommand

import (
	"time"

	"github.com/johtani/smarthome/subcommand/action"
	"github.com/johtani/smarthome/subcommand/action/owntone"
	"github.com/johtani/smarthome/subcommand/action/yamaha"
)

// StartMusicCmd is the command name for starting music playback.
const StartMusicCmd = "start music"

// NewStartMusicCmdDefinition creates the definition for the start music command.
func NewStartMusicCmdDefinition() Definition {
	return Definition{
		Name:        StartMusicCmd,
		Description: "Play a random music selection only when no artist, song, album, or playlist is specified",
		Factory:     NewStartMusicSubcommand,
		Args: []Arg{
			{"mode", "random target type", false, []string{"artist", "genre"}, ""},
		},
	}
}

// NewStartMusicSubcommand creates a new Subcommand for the start music command.
func NewStartMusicSubcommand(definition Definition, config Config) Subcommand {
	owntoneClient := owntone.NewClient(config.Owntone)
	yamahaClient := yamaha.NewClient(config.Yamaha)
	return Subcommand{
		Definition: definition,
		actions: []action.Action{
			owntone.NewClearQueueAction(owntoneClient),
			yamaha.NewPowerOnAction(yamahaClient),
			yamaha.NewSetInputAction(yamahaClient, "airplay"),
			owntone.NewPlayAction(owntoneClient),
			action.NewNoOpAction(3 * time.Second),
			yamaha.NewSetVolumeAction(yamahaClient, 39),
			owntone.NewDisplayOutputsAction(owntoneClient, true),
		},
		ignoreError: true,
	}
}
