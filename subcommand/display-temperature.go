package subcommand

import (
	"github.com/johtani/smarthome/subcommand/action"
	"github.com/johtani/smarthome/subcommand/action/switchbot"
)

const DisplayTemperature = "display-temperature"
const DispTemp = "disp-temp"

func NewDisplayTemperatureDefinition() Definition {
	return Definition{
		DisplayTemperature,
		"Display temperature/humidity",
		NewDisplayTemperatureSubcommnad,
	}
}

func NewDisplayTemperatureSubcommnad(definition Definition, config Config) Subcommand {
	switchbotClient := switchbot.NewClient(config.Switchbot)
	return Subcommand{
		definition,
		[]action.Action{
			switchbot.NewGetTemperatureAndHumidityAction(switchbotClient),
		},
		true,
	}
}
