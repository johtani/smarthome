package subcommand

import (
	"github.com/johtani/smarthome/subcommand/action"
	"github.com/johtani/smarthome/subcommand/action/switchbot"
	switchbotsdk "github.com/nasa9084/go-switchbot/v3"
)

const LightOffCmd = "light off"

func NewLightOffDefinition() Definition {
	return Definition{
		Name:        LightOffCmd,
		Description: "Light off via SwitchBot",
		Factory:     NewLightOffSubcommand,
	}
}

func NewLightOffSubcommand(definition Definition, config Config) Subcommand {
	switchbotClient := switchbot.NewClient(config.Switchbot)
	return Subcommand{
		Definition: definition,
		actions: []action.Action{
			switchbot.NewSendCommandAction(switchbotClient, config.Switchbot.LightDeviceId, switchbotsdk.TurnOffCommand()),
		},
		ignoreError: true,
	}
}
