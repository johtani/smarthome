package subcommand

import (
	"github.com/johtani/smarthome/subcommand/action"
	"github.com/johtani/smarthome/subcommand/action/switchbot"
	switchbotsdk "github.com/nasa9084/go-switchbot/v3"
)

const LightOnCmd = "light-on"

func NewLightOnDefinition() Definition {
	return Definition{
		Name:        LightOnCmd,
		Description: "Light on via SwitchBot",
		Factory:     NewLightOnSubcommand,
	}
}

func NewLightOnSubcommand(definition Definition, config Config) Subcommand {
	switchbotClient := switchbot.NewClient(config.Switchbot)
	return Subcommand{
		Definition: definition,
		actions: []action.Action{
			switchbot.NewSendCommandAction(switchbotClient, config.Switchbot.LightDeviceId, switchbotsdk.TurnOnCommand()),
			switchbot.NewExecuteSceneAction(switchbotClient, config.Switchbot.LightSceneId),
		},
		ignoreError: true,
	}
}
