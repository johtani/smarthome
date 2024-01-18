package subcommand

import (
	"github.com/johtani/smarthome/subcommand/action"
	"github.com/johtani/smarthome/subcommand/action/switchbot"
)

const DisplayTemperatureCmd = "display-temperature"
const DispTempCmd = "disp-temp"

func NewDisplayTemperatureDefinition() Definition {
	return Definition{
		Name:        DisplayTemperatureCmd,
		Description: "Display temperature/humidity",
		Factory:     NewDisplayTemperatureSubcommnad,
	}
}

func NewDisplayTemperatureSubcommnad(definition Definition, config Config) Subcommand {
	switchbotClient := switchbot.NewClient(config.Switchbot)
	return Subcommand{
		Definition: definition,
		actions: []action.Action{
			switchbot.NewGetTemperatureAndHumidityAction(switchbotClient),
		},
		ignoreError: true,
	}
}
