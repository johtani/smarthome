package subcommand

import (
	"os"
	"smart_home/subcommand/action"
	"smart_home/subcommand/action/owntone"
	"smart_home/subcommand/action/switchbot"
)

const StartMeetingCmd = "start-meeting"
const EnvStartMeetingScene = "SWITCHBOT_START_MEETING_SCENE"

func NewStartMeetingSubcommand() Subcommand {
	return Subcommand{
		StartMeetingCmd,
		"Actions before starting meeting",
		[]action.Action{
			owntone.NewPauseAction(),
			switchbot.NewExecuteSceneAction(os.Getenv(EnvStartMeetingScene)),
		},
		checkConfig,
	}
}
