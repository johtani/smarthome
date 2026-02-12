package subcommand

import (
	"github.com/johtani/smarthome/subcommand/action"
	"github.com/johtani/smarthome/subcommand/action/switchbot"
	switchbotsdk "github.com/nasa9084/go-switchbot/v3"
)

// AirConditionerOffCmd is the command name for turning off the air conditioner.
const AirConditionerOffCmd = "air conditioner off"

// ACOffCmd is a short name for the air conditioner off command.
const ACOffCmd = "ac off"

// NewAirConditionerOffDefinition creates the definition for the air conditioner off command.
func NewAirConditionerOffDefinition() Definition {
	return Definition{
		Name:        AirConditionerOffCmd,
		Description: "Air Conditioner switch off via SwitchBot",
		Factory:     NewAirConditionerOffSubcommand,
		shortnames:  []string{ACOffCmd},
	}
}

// NewAirConditionerOffSubcommand creates a new Subcommand for the air conditioner off command.
func NewAirConditionerOffSubcommand(definition Definition, config Config) Subcommand {
	switchbotClient := switchbot.NewClient(config.Switchbot)
	return Subcommand{
		Definition: definition,
		actions: []action.Action{
			switchbot.NewSendCommandAction(switchbotClient, config.Switchbot.AirConditionerID, switchbotsdk.TurnOffCommand()),
		},
		ignoreError: true,
	}
}
