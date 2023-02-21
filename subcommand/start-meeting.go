package subcommand

import (
	"smart_home/subcommand/action"
	"smart_home/subcommand/action/owntone"
)

const StartMeetingCmd = "start-meeting"

func NewStartMeetingSubcommand() Subcommand {
	return Subcommand{
		StartMeetingCmd,
		"Actions before starting meeting",
		[]action.Action{
			owntone.NewPauseAction(),
		},
		checkConfig,
	}
}
