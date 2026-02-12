package subcommand

import (
	"github.com/johtani/smarthome/subcommand/action"
	"github.com/johtani/smarthome/subcommand/action/switchbot"
	switchbotsdk "github.com/nasa9084/go-switchbot/v3"
)

// AirConditionerOnCmd is the command name for turning on the air conditioner.
const AirConditionerOnCmd = "air conditioner on"

// ACOnCmd is a short name for the air conditioner on command.
const ACOnCmd = "ac on"

// NewAirConditionerOnDefinition creates the definition for the air conditioner on command.
func NewAirConditionerOnDefinition() Definition {
	return Definition{
		Name:        AirConditionerOnCmd,
		Description: "Air Conditioner switch on via SwitchBot",
		Factory:     NewAirConditionerOnSubcommand,
		shortnames:  []string{ACOnCmd},
	}
}

// NewAirConditionerOnSubcommand creates a new Subcommand for the air conditioner on command.
func NewAirConditionerOnSubcommand(definition Definition, config Config) Subcommand {
	switchbotClient := switchbot.NewClient(config.Switchbot)
	return Subcommand{
		Definition: definition,
		actions: []action.Action{
			switchbot.NewSendCommandAction(switchbotClient, config.Switchbot.AirConditionerID, switchbotsdk.TurnOnCommand()),
		},
		ignoreError: true,
	}
}
