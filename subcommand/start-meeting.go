package subcommand

import (
	"os"
	"smart_home/subcommand/action"
	"smart_home/subcommand/action/owntone"
	"smart_home/subcommand/action/switchbot"
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
	owntoneClient := owntone.NewOwntoneClient(config.owntone)
	switchbotClient := switchbot.NewSwitchBotClient(config.switchbot)
	return Subcommand{
		definition,
		[]action.Action{
			owntone.NewPauseAction(owntoneClient),
			switchbot.NewExecuteSceneAction(switchbotClient, os.Getenv(EnvStartMeetingScene)),
		},
		true,
	}
}
