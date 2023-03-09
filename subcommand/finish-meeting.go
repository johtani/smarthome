package subcommand

import (
	"smart_home/subcommand/action"
	"smart_home/subcommand/action/owntone"
)

const FinishMeetingCmd = "finish-meeting"

func NewFinishMeetingSubcommand() Subcommand {
	return Subcommand{
		FinishMeetingCmd,
		"Actions after meeting",
		[]action.Action{
			owntone.NewPlayAction(),
			owntone.NewSetVolumeAction(),
		},
		checkConfig,
		true,
	}
}
