package subcommand

import (
	"github.com/johtani/smarthome/subcommand/action"
	"github.com/johtani/smarthome/subcommand/action/switchbot"
	switchbot2 "github.com/nasa9084/go-switchbot/v2"
	"os"
)

const LightOffCmd = "light-off"
const EnvLightDeviceId = "SWITCHBOT_LIGHT_DEVICE_ID"

func NewLightOffDefinition() Definition {
	return Definition{
		LightOffCmd,
		"Light off via SwitchBot",
		NewLightOffSubcommand,
	}
}

func NewLightOffSubcommand(definition Definition, config Config) Subcommand {
	switchbotClient := switchbot.NewClient(config.switchbot)
	return Subcommand{
		definition,
		[]action.Action{
			switchbot.NewSendCommandAction(switchbotClient, os.Getenv(EnvLightDeviceId), switchbot2.TurnOffCommand()),
		},
		true,
	}
}
