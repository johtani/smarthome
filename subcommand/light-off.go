package subcommand

import (
	"github.com/johtani/smarthome/subcommand/action"
	"github.com/johtani/smarthome/subcommand/action/switchbot"
	switchbotsdk "github.com/nasa9084/go-switchbot/v3"
)

// LightOffCmd is the command name for turning off the light.
const LightOffCmd = "light off"

// NewLightOffDefinition creates the definition for the light off command.
func NewLightOffDefinition() Definition {
	return Definition{
		Name:        LightOffCmd,
		Description: "Light off via SwitchBot",
		Factory:     NewLightOffSubcommand,
	}
}

// NewLightOffSubcommand creates a new Subcommand for the light off command.
func NewLightOffSubcommand(definition Definition, config Config) Subcommand {
	switchbotClient := switchbot.NewClient(config.Switchbot)
	return Subcommand{
		Definition: definition,
		actions: []action.Action{
			switchbot.NewSendCommandAction(switchbotClient, config.Switchbot.LightDeviceID, switchbotsdk.TurnOffCommand()),
		},
		ignoreError: true,
	}
}
