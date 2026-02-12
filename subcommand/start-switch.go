package subcommand

import (
	"time"

	"github.com/johtani/smarthome/subcommand/action"
	"github.com/johtani/smarthome/subcommand/action/yamaha"
)

// StartSwitchCmd is the command name for starting the Nintendo Switch setup.
const StartSwitchCmd = "start switch"

// NewStartSwitchDefinition creates the definition for the start switch command.
func NewStartSwitchDefinition() Definition {
	return Definition{
		Name:        StartSwitchCmd,
		Description: "Actions before starting switch",
		Factory:     NewStartSwitchSubcommand,
	}
}

// NewStartSwitchSubcommand creates a new Subcommand for the start switch command.
func NewStartSwitchSubcommand(definition Definition, config Config) Subcommand {
	yamahaClient := yamaha.NewClient(config.Yamaha)
	return Subcommand{
		Definition: definition,
		actions: []action.Action{
			yamaha.NewSetSceneAction(yamahaClient, 2),
			action.NewNoOpAction(1 * time.Second),
			yamaha.NewSetVolumeAction(yamahaClient, 70),
		},
		ignoreError: true,
	}
}
