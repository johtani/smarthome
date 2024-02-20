package subcommand

import (
	"github.com/johtani/smarthome/subcommand/action"
	"github.com/johtani/smarthome/subcommand/action/owntone"
	"time"
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
	return Subcommand{
		Definition: definition,
		actions: []action.Action{
			owntone.NewPlayAction(owntoneClient),
			action.NewNoOpAction(3 * time.Second),
			owntone.NewSetVolumeAction(owntoneClient),
		},
		ignoreError: true,
	}
}
