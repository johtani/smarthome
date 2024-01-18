package subcommand

import (
	"github.com/johtani/smarthome/subcommand/action"
	"github.com/johtani/smarthome/subcommand/action/switchbot"
	switchbotsdk "github.com/nasa9084/go-switchbot/v3"
)

const AirConditionerOnCmd = "air-conditioner-on"
const ACOnCmd = "ac-on"

func NewAirConditionerOnDefinition() Definition {
	return Definition{
		Name:        AirConditionerOnCmd,
		Description: "Air Conditioner switch on via SwitchBot",
		Factory:     NewAirConditionerOnSubcommand,
	}
}

func NewAirConditionerOnSubcommand(definition Definition, config Config) Subcommand {
	switchbotClient := switchbot.NewClient(config.Switchbot)
	return Subcommand{
		Definition: definition,
		actions: []action.Action{
			switchbot.NewSendCommandAction(switchbotClient, config.Switchbot.AirConditionerId, switchbotsdk.TurnOnCommand()),
		},
		ignoreError: true,
	}
}
