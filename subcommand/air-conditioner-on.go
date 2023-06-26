package subcommand

import (
	"github.com/johtani/smarthome/subcommand/action"
	"github.com/johtani/smarthome/subcommand/action/switchbot"
	switchbot2 "github.com/nasa9084/go-switchbot/v2"
)

const AirConditionerOnCmd = "air-conditioner-on"
const ACOnCmd = "ac-on"

func NewAirConditionerOnDefinition() Definition {
	return Definition{
		AirConditionerOnCmd,
		"Air Conditioner switch on via SwitchBot",
		NewAirConditionerOnSubcommand,
	}
}

func NewAirConditionerOnSubcommand(definition Definition, config Config) Subcommand {
	switchbotClient := switchbot.NewClient(config.Switchbot)
	return Subcommand{
		definition,
		[]action.Action{
			switchbot.NewSendCommandAction(switchbotClient, config.Switchbot.AirConditionerId, switchbot2.TurnOnCommand()),
		},
		true,
	}
}
