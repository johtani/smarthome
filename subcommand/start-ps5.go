package subcommand

import (
	"github.com/johtani/smarthome/subcommand/action"
	"github.com/johtani/smarthome/subcommand/action/yamaha"
	"time"
)

const StartPS5Cmd = "start ps5"

func NewStartPS5Definition() Definition {
	return Definition{
		Name:        StartPS5Cmd,
		Description: "Actions before starting PS5",
		Factory:     NewStartPS5Subcommand,
	}
}

func NewStartPS5Subcommand(definition Definition, config Config) Subcommand {
	yamahaClient := yamaha.NewClient(config.Yamaha)
	return Subcommand{
		Definition: definition,
		actions: []action.Action{
			yamaha.NewSetSceneAction(yamahaClient, 1),
			action.NewNoOpAction(1 * time.Second),
			yamaha.NewSetVolumeAction(yamahaClient),
		},
		ignoreError: true,
	}
}
