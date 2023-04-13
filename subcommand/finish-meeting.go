package subcommand

import (
	"smart_home/subcommand/action"
	"smart_home/subcommand/action/owntone"
	"time"
)

const FinishMeetingCmd = "finish-meeting"

func NewFinishMeetingSubcommand(config Config) Subcommand {
	owntoneClient := owntone.NewOwntoneClient(config.owntone)
	return Subcommand{
		FinishMeetingCmd,
		"Actions after meeting",
		[]action.Action{
			owntone.NewPlayAction(owntoneClient),
			action.NewNoOpAction(3 * time.Second),
			owntone.NewSetVolumeAction(owntoneClient),
		},
		true,
	}
}
