package subcommand

import (
	"time"

	"github.com/johtani/smarthome/subcommand/action"
	"github.com/johtani/smarthome/subcommand/action/yamaha"
)

const StartSwitchCmd = "start switch"

func NewStartSwitchDefinition() Definition {
	return Definition{
		Name:        StartSwitchCmd,
		Description: "Actions before starting switch",
		Factory:     NewStartSwitchSubcommand,
	}
}

func NewStartSwitchSubcommand(definition Definition, config Config) Subcommand {
	yamahaClient := yamaha.NewClient(config.Yamaha)
	return Subcommand{
		Definition: definition,
		actions: []action.Action{
			yamaha.NewSetSceneAction(yamahaClient, 2),
			action.NewNoOpAction(1 * time.Second),
			yamaha.NewSetVolumeAction(yamahaClient),
		},
		ignoreError: true,
	}
}
