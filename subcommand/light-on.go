package subcommand

import (
	switchbot2 "github.com/nasa9084/go-switchbot/v2"
	"os"
	"smarthome/subcommand/action"
	"smarthome/subcommand/action/switchbot"
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
