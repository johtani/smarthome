package subcommand

import (
	"github.com/johtani/smarthome/subcommand/action"
	"github.com/johtani/smarthome/subcommand/action/owntone"
	"time"
)

const StartMusicCmd = "start music"

func NewStartMusicCmdDefinition() Definition {
	return Definition{
		Name:        StartMusicCmd,
		Description: "Start Music from playlist or by artist",
		Factory:     NewStartMusicSubcommand,
	}
}

func NewStartMusicSubcommand(definition Definition, config Config) Subcommand {
	owntoneClient := owntone.NewClient(config.Owntone)
	return Subcommand{
		Definition: definition,
		actions: []action.Action{
			owntone.NewClearQueueAction(owntoneClient),
			owntone.NewPlayAction(owntoneClient),
			action.NewNoOpAction(3 * time.Second),
			owntone.NewSetVolumeAction(owntoneClient),
		},
		ignoreError: true,
	}
}
