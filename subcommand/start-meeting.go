package subcommand

import (
	"github.com/johtani/smarthome/subcommand/action"
	"github.com/johtani/smarthome/subcommand/action/owntone"
	"github.com/johtani/smarthome/subcommand/action/switchbot"
	"github.com/johtani/smarthome/subcommand/action/yamaha"
)

const StartMeetingCmd = "start-meeting"

func NewStartMeetingDefinition() Definition {
	return Definition{
		StartMeetingCmd,
		"Actions before starting meeting",
		NewStartMeetingSubcommand,
	}
}

func NewStartMeetingSubcommand(definition Definition, config Config) Subcommand {
	owntoneClient := owntone.NewClient(config.Owntone)
	switchbotClient := switchbot.NewClient(config.Switchbot)
	yamahaClient := yamaha.NewClient(config.Yamaha)
	return Subcommand{
		definition,
		[]action.Action{
			owntone.NewPauseAction(owntoneClient),
			switchbot.NewExecuteSceneAction(switchbotClient, config.Switchbot.LightSceneId),
			yamaha.NewPowerOffAction(yamahaClient),
		},
		true,
	}
}
