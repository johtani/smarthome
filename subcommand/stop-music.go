package subcommand

import (
	"github.com/johtani/smarthome/subcommand/action"
	"github.com/johtani/smarthome/subcommand/action/owntone"
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
	return Subcommand{
		definition,
		[]action.Action{
			owntone.NewPauseAction(owntoneClient),
		},
		true,
	}
}
