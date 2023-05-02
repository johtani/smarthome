package subcommand

import (
	"github.com/johtani/smarthome/subcommand/action"
	"github.com/johtani/smarthome/subcommand/action/switchbot"
	switchbot2 "github.com/nasa9084/go-switchbot/v2"
	"os"
)

const LightOnCmd = "light-on"

func NewLightOnDefinition() Definition {
	return Definition{
		LightOnCmd,
		"Light on via SwitchBot",
		NewLightOnSubcommand,
	}
}

func NewLightOnSubcommand(definition Definition, config Config) Subcommand {
	switchbotClient := switchbot.NewClient(config.switchbot)
	return Subcommand{
		definition,
		[]action.Action{
			switchbot.NewSendCommandAction(switchbotClient, os.Getenv(EnvLightDeviceId), switchbot2.TurnOnCommand()),
			switchbot.NewExecuteSceneAction(switchbotClient, os.Getenv(EnvStartMeetingScene)),
		},
		true,
	}
}
