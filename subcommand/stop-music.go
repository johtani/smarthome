package subcommand

import (
	"github.com/johtani/smarthome/subcommand/action"
	"github.com/johtani/smarthome/subcommand/action/owntone"
	"github.com/johtani/smarthome/subcommand/action/yamaha"
)

const StopMusicCmd = "stop-music"

func NewStopMusicDefinition() Definition {
	return Definition{
		StopMusicCmd,
		"Stop Music",
		NewStopMusicSubcommand,
	}
}

func NewStopMusicSubcommand(definition Definition, config Config) Subcommand {
	owntoneClient := owntone.NewClient(config.Owntone)
	yamahaClient := yamaha.NewClient(config.Yamaha)
	return Subcommand{
		definition,
		[]action.Action{
			owntone.NewPauseAction(owntoneClient),
			yamaha.NewPowerOffAction(yamahaClient),
		},
		true,
	}
}
