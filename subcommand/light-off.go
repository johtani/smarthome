package subcommand

import (
	"github.com/johtani/smarthome/subcommand/action"
	"github.com/johtani/smarthome/subcommand/action/switchbot"
	switchbot2 "github.com/nasa9084/go-switchbot/v2"
)

const LightOffCmd = "light-off"

func NewLightOffDefinition() Definition {
	return Definition{
		LightOffCmd,
		"Light off via SwitchBot",
		NewLightOffSubcommand,
	}
}

func NewLightOffSubcommand(definition Definition, config Config) Subcommand {
	switchbotClient := switchbot.NewClient(config.Switchbot)
	return Subcommand{
		definition,
		[]action.Action{
			switchbot.NewSendCommandAction(switchbotClient, config.Switchbot.LightDeviceId, switchbot2.TurnOffCommand()),
		},
		true,
	}
}
