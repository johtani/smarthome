package subcommand

import (
	switchbot2 "github.com/nasa9084/go-switchbot/v2"
	"os"
	"smart_home/subcommand/action"
	"smart_home/subcommand/action/switchbot"
)

const LightOffCmd = "light-off"
const EnvLightDeviceId = "SWITCHBOT_LIGHT_DEVICE_ID"

func NewLightOffDefinition() Definition {
	return Definition{
		LightOffCmd,
		"List devices on SwitchBot",
		NewLightOffSubcommand,
	}
}

func NewLightOffSubcommand(definition Definition, config Config) Subcommand {
	switchbotClient := switchbot.NewSwitchBotClient(config.switchbot)
	return Subcommand{
		definition,
		[]action.Action{
			switchbot.NewSendCommandAction(switchbotClient, os.Getenv(EnvLightDeviceId), switchbot2.TurnOffCommand()),
		},
		true,
	}
}
