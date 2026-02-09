package subcommand

import (
	"time"

	"github.com/johtani/smarthome/subcommand/action"
	"github.com/johtani/smarthome/subcommand/action/owntone"
	"github.com/johtani/smarthome/subcommand/action/yamaha"
)

const FinishMeetingCmd = "finish meeting"

func NewFinishMeetingDefinition() Definition {
	return Definition{
		Name:        FinishMeetingCmd,
		Description: "Actions after meeting",
		Factory:     NewFinishMeetingSubcommand,
	}
}

func NewFinishMeetingSubcommand(definition Definition, config Config) Subcommand {
	owntoneClient := owntone.NewClient(config.Owntone)
	yamahaClient := yamaha.NewClient(config.Yamaha)
	return Subcommand{
		Definition: definition,
		actions: []action.Action{
			yamaha.NewPowerOnAction(yamahaClient),
			yamaha.NewSetInputAction(yamahaClient, "airplay"),
			owntone.NewPlayAction(owntoneClient),
			action.NewNoOpAction(3 * time.Second),
			yamaha.NewSetVolumeAction(yamahaClient, 39),
		},
		ignoreError: true,
	}
}
