package subcommand

import (
	switchbot2 "github.com/nasa9084/go-switchbot/v2"
	"os"
	"smart_home/subcommand/action"
	"smart_home/subcommand/action/switchbot"
)

const LightOffCmd = "light-off"
const EnvLightDeviceId = "SWITCHBOT_LIGHT_DEVICE_ID"

func NewLightOffSubcommand(config Config) Subcommand {
	switchbotClient := switchbot.NewSwitchBotClient(config.switchbot)
	return Subcommand{
		LightOffCmd,
		"List devices on SwitchBot",
		[]action.Action{
			switchbot.NewSendCommandAction(switchbotClient, os.Getenv(EnvLightDeviceId), switchbot2.TurnOffCommand()),
		},
		true,
	}
}
