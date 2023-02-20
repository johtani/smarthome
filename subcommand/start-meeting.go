package subcommand

import (
	"smart_home/subcommand/action"
	"smart_home/subcommand/action/owntone"
)

func checkConfig() error {
	// TODO エラーはまとめて返したほうがいいかも
	err := owntone.CheckConfig()
	if err != nil {
		return err
	}
	return nil
}

const START_MEETING_CMD = "start-meeting"

func NewStartMeetingSubcommand() Subcommand {
	return Subcommand{
		START_MEETING_CMD,
		"Actions before starting meeting",
		[]action.Action{
			owntone.NewPauseAction(),
		},
		checkConfig,
	}
}
