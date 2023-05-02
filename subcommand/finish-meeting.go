package subcommand

import (
	"github.com/johtani/smarthome/subcommand/action"
	"github.com/johtani/smarthome/subcommand/action/owntone"
	"time"
)

const FinishMeetingCmd = "finish-meeting"

func NewFinishMeetingDefinition() Definition {
	return Definition{
		FinishMeetingCmd,
		"Actions after meeting",
		NewFinishMeetingSubcommand,
	}
}

func NewFinishMeetingSubcommand(definition Definition, config Config) Subcommand {
	owntoneClient := owntone.NewClient(config.owntone)
	return Subcommand{
		definition,
		[]action.Action{
			owntone.NewPlayAction(owntoneClient),
			action.NewNoOpAction(3 * time.Second),
			owntone.NewSetVolumeAction(owntoneClient),
		},
		true,
	}
}
