package subcommand

import (
	"github.com/johtani/smarthome/subcommand/action"
	"github.com/johtani/smarthome/subcommand/action/owntone"
	"github.com/johtani/smarthome/subcommand/action/yamaha"
)

// StopMusicCmd is the command name for stopping music playback.
const StopMusicCmd = "stop music"

// NewStopMusicDefinition creates the definition for the stop music command.
func NewStopMusicDefinition() Definition {
	return Definition{
		Name:        StopMusicCmd,
		Description: "Stop Music",
		Factory:     NewStopMusicSubcommand,
	}
}

// NewStopMusicSubcommand creates a new Subcommand for the stop music command.
func NewStopMusicSubcommand(definition Definition, config Config) Subcommand {
	owntoneClient := owntone.NewClient(config.Owntone)
	yamahaClient := yamaha.NewClient(config.Yamaha)
	return Subcommand{
		Definition: definition,
		actions: []action.Action{
			owntone.NewPauseAction(owntoneClient),
			yamaha.NewPowerOffAction(yamahaClient),
		},
		ignoreError: true,
	}
}
