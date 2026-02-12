package subcommand

import (
	"time"

	"github.com/johtani/smarthome/subcommand/action"
	"github.com/johtani/smarthome/subcommand/action/yamaha"
)

// StartPS5Cmd is the command name for starting the PS5 setup.
const StartPS5Cmd = "start ps5"

// NewStartPS5Definition creates the definition for the start ps5 command.
func NewStartPS5Definition() Definition {
	return Definition{
		Name:        StartPS5Cmd,
		Description: "Actions before starting PS5",
		Factory:     NewStartPS5Subcommand,
	}
}

// NewStartPS5Subcommand creates a new Subcommand for the start ps5 command.
func NewStartPS5Subcommand(definition Definition, config Config) Subcommand {
	yamahaClient := yamaha.NewClient(config.Yamaha)
	return Subcommand{
		Definition: definition,
		actions: []action.Action{
			yamaha.NewSetSceneAction(yamahaClient, 1),
			action.NewNoOpAction(1 * time.Second),
			yamaha.NewSetVolumeAction(yamahaClient, 70),
		},
		ignoreError: true,
	}
}
