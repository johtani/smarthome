package subcommand

import (
	"os"
	"smarthome/subcommand/action"
	"smarthome/subcommand/action/owntone"
	"smarthome/subcommand/action/switchbot"
)

const StartMeetingCmd = "start-meeting"
const EnvStartMeetingScene = "SWITCHBOT_START_MEETING_SCENE"

func NewStartMeetingDefinition() Definition {
	return Definition{
		StartMeetingCmd,
		"Actions before starting meeting",
		NewStartMeetingSubcommand,
	}
}

func NewStartMeetingSubcommand(definition Definition, config Config) Subcommand {
	owntoneClient := owntone.NewClient(config.owntone)
	switchbotClient := switchbot.NewClient(config.switchbot)
	return Subcommand{
		definition,
		[]action.Action{
			owntone.NewPauseAction(owntoneClient),
			switchbot.NewExecuteSceneAction(switchbotClient, os.Getenv(EnvStartMeetingScene)),
		},
		true,
	}
}
