package subcommand

import (
	"github.com/johtani/smarthome/subcommand/action"
	"github.com/johtani/smarthome/subcommand/action/switchbot"
	switchbotsdk "github.com/nasa9084/go-switchbot/v3"
)

// LightOnCmd is the command name for turning on the light.
const LightOnCmd = "light on"

// NewLightOnDefinition creates the definition for the light on command.
func NewLightOnDefinition() Definition {
	return Definition{
		Name:        LightOnCmd,
		Description: "Light on via SwitchBot",
		Factory:     NewLightOnSubcommand,
	}
}

// NewLightOnSubcommand creates a new Subcommand for the light on command.
func NewLightOnSubcommand(definition Definition, config Config) Subcommand {
	switchbotClient := switchbot.NewClient(config.Switchbot)
	return Subcommand{
		Definition: definition,
		actions: []action.Action{
			switchbot.NewSendCommandAction(switchbotClient, config.Switchbot.LightDeviceID, switchbotsdk.TurnOnCommand()),
			switchbot.NewExecuteSceneAction(switchbotClient, config.Switchbot.LightSceneID),
		},
		ignoreError: true,
	}
}
