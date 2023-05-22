package subcommand

import (
	"github.com/johtani/smarthome/subcommand/action"
	"github.com/johtani/smarthome/subcommand/action/yamaha"
	"time"
)

const StartSwitchCmd = "start-switch"

func NewStartSwitchDefinition() Definition {
	return Definition{
		StartSwitchCmd,
		"Actions before starting switch",
		NewStartSwitchSubcommand,
	}
}

func NewStartSwitchSubcommand(definition Definition, config Config) Subcommand {
	yamahaClient := yamaha.NewClient(config.yamaha)
	return Subcommand{
		definition,
		[]action.Action{
			yamaha.NewSetSceneAction(yamahaClient, 2),
			action.NewNoOpAction(1 * time.Second),
			yamaha.NewSetVolumeAction(yamahaClient),
		},
		true,
	}
}
