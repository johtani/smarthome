package subcommand

import (
	"github.com/johtani/smarthome/subcommand/action"
	"github.com/johtani/smarthome/subcommand/action/switchbot"
	switchbotsdk "github.com/nasa9084/go-switchbot/v3"
)

const AirConditionerOffCmd = "air-conditioner-off"
const ACOffCmd = "ac-off"

func NewAirConditionerOffDefinition() Definition {
	return Definition{
		Name:        AirConditionerOffCmd,
		Description: "Air Conditioner switch off via SwitchBot",
		Factory:     NewAirConditionerOffSubcommand,
		shortnames:  []string{ACOffCmd},
	}
}

func NewAirConditionerOffSubcommand(definition Definition, config Config) Subcommand {
	switchbotClient := switchbot.NewClient(config.Switchbot)
	return Subcommand{
		Definition: definition,
		actions: []action.Action{
			switchbot.NewSendCommandAction(switchbotClient, config.Switchbot.AirConditionerId, switchbotsdk.TurnOffCommand()),
		},
		ignoreError: true,
	}
}
