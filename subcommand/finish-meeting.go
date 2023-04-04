package subcommand

import (
	"smart_home/subcommand/action"
	"smart_home/subcommand/action/owntone"
	"time"
)

const FinishMeetingCmd = "finish-meeting"

func NewFinishMeetingSubcommand() Subcommand {
	return Subcommand{
		FinishMeetingCmd,
		"Actions after meeting",
		[]action.Action{
			owntone.NewPlayAction(),
			action.NewNoOpAction(3 * time.Second),
			owntone.NewSetVolumeAction(),
		},
		checkConfig,
		true,
	}
}
